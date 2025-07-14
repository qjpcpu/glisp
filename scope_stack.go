package glisp

import (
	"errors"
	"fmt"
	"sync"
)

// Scope defines a single layer of lexical bindings, mapping symbol numbers to their Sexp values.
type Scope map[int]Sexp

// ScopeStack manages the hierarchy of lexical scopes. It is implemented as a
// stack of ScopeLayers.
//
// Its core design enables efficient and safe lexical scoping for closures through
// a combination of reference counting and a copy-on-write (COW) strategy.
//
//   - Reference Counting: Each ScopeLayer has a reference count (`ref`). When a
//     new environment (e.g., for a closure) is created via Fork(), it shares the
//     existing layers and increments their reference counts.
//
//   - Copy-on-Write: When a modification is attempted on a shared layer (ref > 1),
//     a new layer is created and pushed onto the stack for the current environment
//     instead of modifying the shared layer directly. This ensures that changes in one
//     environment do not affect others that share the same parent scope.
type ScopeStack struct {
	top, bottom *ScopeLayer
}

// ScopeLayer represents a single frame in the scope stack.
type ScopeLayer struct {
	Scope Scope
	// ref is the reference count. A layer can be shared by multiple ScopeStacks
	// (e.g., a parent function's scope and a closure's scope). This count
	// tracks how many stacks are currently referencing this layer.
	// The layer is only recycled when ref drops to 0.
	ref int
	// next points to the parent scope layer in the stack.
	next *ScopeLayer
}

// NewScopeStack creates and returns an empty ScopeStack.
func NewScopeStack() *ScopeStack {
	return &ScopeStack{}
}

// newScope retrieves a Scope map from the object pool.
func newScope() Scope { return scopePool.Get().(Scope) }

// newScopeLayer retrieves a ScopeLayer from the object pool and initializes it
// with a new, empty Scope.
func newScopeLayer() *ScopeLayer {
	return newScopeLayerWith(newScope())
}

// newScopeLayerWith retrieves a ScopeLayer from the object pool and initializes it
// with the provided Scope.
func newScopeLayerWith(s Scope) *ScopeLayer {
	if s == nil {
		s = newScope()
	}
	layer := scopeLayerPool.Get().(*ScopeLayer)
	layer.ref = 1
	layer.Scope = s
	return layer
}

// IsStackElem is a marker method for the StackElem interface.
func (s *ScopeLayer) IsStackElem() {}

// Clone creates a deep copy of the ScopeLayer and its underlying Scope map.
func (s *ScopeLayer) Clone() *ScopeLayer { // newScopeLayer() ref is 1
	newScope := newScopeLayer()
	for k, v := range s.Scope {
		newScope.Scope[k] = v
	}
	return newScope
}

// Find searches for a symbol's value within this specific scope layer.
func (s *ScopeLayer) Find(n int) (Sexp, bool) {
	if expr, ok := s.Scope[n]; ok {
		return expr, true
	}
	return SexpNull, false
}

// Bind sets the value for a symbol in this specific scope layer.
func (s *ScopeLayer) Bind(n int, e Sexp) {
	s.Scope[n] = e
}

// incrRef increments the reference count of all layers from the given 'top'
// layer down to the bottom of the stack. This is called when a stack is forked.
func (stack *ScopeStack) incrRef(top *ScopeLayer) {
	for ptr := top; ptr != nil; {
		ptr.ref++
		ptr = ptr.next
	}
}

// Fork creates a new ScopeStack that shares the same underlying layers
// with the original stack. It increments the reference count of the shared
// layers. This is the primary mechanism for creating an execution environment
// for a closure, efficiently capturing its lexical scope.
func (stack *ScopeStack) Fork() *ScopeStack {
	st := NewScopeStack()
	st.top = stack.top
	st.bottom = stack.bottom
	stack.incrRef(stack.top)
	return st
}

// ForkBottom creates a new ScopeStack that starts from the global scope.
// (the bottom layer) of the original stack. This is used for standard function
// calls that do not capture a closure but need access to global bindings.
func (stack *ScopeStack) ForkBottom() *ScopeStack {
	st := NewScopeStack()
	st.top = stack.bottom
	st.bottom = stack.bottom
	stack.incrRef(stack.bottom)
	return st
}

// IsStackElem is a marker method for the StackElem interface.
func (stack *ScopeStack) IsStackElem() {}

// Clone creates a full, deep copy of the ScopeStack, including all its layers.
// This is a more expensive operation than Fork and is used when a completely
// independent environment is required.
func (stack *ScopeStack) Clone() *ScopeStack {
	stack2 := NewScopeStack()
	var prev2 *ScopeLayer
	for ptr := stack.top; ptr != nil; {
		ptr2 := ptr.Clone()
		if prev2 == nil {
			stack2.top = ptr2
		} else {
			prev2.next = ptr2
		}
		prev2 = ptr2
		if ptr.next == nil {
			stack2.bottom = ptr2
		}
		ptr = ptr.next
	}
	return stack2
}

// push adds a new layer to the top of the stack.
func (stack *ScopeStack) push(layer *ScopeLayer) {
	if layer == nil {
		return
	}
	if stack.top != nil {
		layer.next = stack.top
	} else {
		stack.bottom = layer
	}
	if layer.ref == 0 {
		layer.ref = 1
	}
	stack.top = layer
}

// Pop removes the top layer from the stack. It decrements the layer's reference
// count and recycles it back to the pool if the count drops to zero.
func (stack *ScopeStack) Pop() error {
	if stack.top == nil {
		return errors.New("pop from empty scope stack")
	}
	cur := stack.top
	cur.ref--
	stack.top = cur.next
	if stack.top == nil {
		stack.bottom = nil
	}
	recycleScopeLayer(cur)
	return nil
}

// PushScope creates a new, empty scope and pushes it onto the stack.
// This is used for constructs like `let`.
func (stack *ScopeStack) PushScope() {
	stack.Push()
}

// PopScope removes the top scope from the stack.
func (stack *ScopeStack) PopScope() error {
	return stack.Pop()
}

// Push adds one or more pre-existing scopes to the top of the stack.
func (stack *ScopeStack) Push(scopes ...Scope) {
	if len(scopes) == 0 {
		layer := newScopeLayer()
		stack.push(layer)
	} else {
		for i := range scopes {
			stack.push(newScopeLayerWith(scopes[i]))
		}
	}
}

// LookupSymbol searches for a symbol starting from the top of the stack and
// moving down through parent scopes until the symbol is found.
func (stack *ScopeStack) LookupSymbol(sym SexpSymbol) (Sexp, error) {
	for ptr := stack.top; ptr != nil; {
		if expr, ok := ptr.Find(sym.number); ok {
			return expr, nil
		}
		ptr = ptr.next
	}
	return SexpNull, fmt.Errorf("symbol `%v` not found", sym.Name())
}

// GlobalFuntions returns a list of function names found in the global scope (bottom layer).
func (stack *ScopeStack) GlobalFuntions() (ret []string) {
	if stack.IsEmpty() {
		return nil
	}
	for _, v := range stack.bottom.Scope {
		if fn, ok := v.(*SexpFunction); ok {
			ret = append(ret, fn.name)
		}
	}
	return
}

// IsEmpty returns true if the stack has no layers.
func (stack *ScopeStack) IsEmpty() bool {
	return stack.top == nil
}

type BindSymbolOption int

const (
	BIND_DEFAULT BindSymbolOption = iota
	BIND_GLOBAL
)

// BindSymbol binds a symbol to a value. This is the implementation for `def`.
// It embodies the "Copy-on-Write" (COW) strategy.
//
// If the binding is global, it binds in the bottom-most scope.
// Otherwise, it checks if the top scope is shared (ref > 1).
// If it is shared, a new scope is pushed to the stack first to avoid
// modifying the shared scope. The binding is then performed in the new,
// unshared top scope. This ensures lexical scope integrity for closures.
func (stack *ScopeStack) BindSymbol(sym SexpSymbol, expr Sexp, options ...BindSymbolOption) error {
	if stack.IsEmpty() {
		return errors.New("no scope available")
	}
	if len(options) > 0 && options[0] == BIND_GLOBAL {
		stack.bottom.Bind(sym.number, expr)
	} else {
		// This is the heart of the Copy-on-Write (COW) strategy.
		//
		// Before binding a symbol in the current scope (stack.top), we check two conditions:
		//
		// 1. `stack.top.ref > 1`: Is this scope layer shared?
		//    A reference count greater than 1 indicates that a closure has captured
		//    this scope. To maintain lexical scope integrity, we must not
		//    mutate this shared layer directly.
		//
		// 2. `stack.top != stack.bottom`: Is this a non-global scope?
		//    We explicitly exclude the global scope from COW, as it is intended
		//    to be a shared, mutable environment.
		//
		// If both are true, we push a new, unshared scope onto the stack before binding.
		// This ensures that the new binding is local to the current environment and
		// does not affect the closure's captured scope.
		if stack.top != stack.bottom && stack.top.ref > 1 {
			stack.PushScope()
		}
		stack.top.Bind(sym.number, expr)
	}
	return nil
}

// SetSymbol updates an existing binding for a symbol. This is the implementation for `set!`.
// It searches up the scope stack for the first occurrence of the symbol and
// modifies its value in-place.
//
// NOTE: By design, this can mutate a shared parent scope. This behavior is
// intentional to mimic `set!` in many Lisp dialects.
func (stack *ScopeStack) SetSymbol(sym SexpSymbol, expr Sexp) error {
	if stack.IsEmpty() {
		return errors.New("no scope available")
	}
	for ptr := stack.top; ptr != nil; {
		if _, ok := ptr.Find(sym.number); ok {
			ptr.Bind(sym.number, expr)
			return nil
		}
		ptr = ptr.next
	}
	return stack.BindSymbol(sym, expr)
}

// Clear removes all layers from the stack, decrementing their reference counts
// and recycling them if they are no longer referenced by any other stack.
// This is critical for releasing resources when an environment is discarded.
func (stack *ScopeStack) Clear() {
	for stack.top != nil {
		cur := stack.top
		cur.ref--
		stack.top = cur.next
		recycleScopeLayer(cur)
	}
	stack.bottom = nil
}

// scopePool holds maps for reuse.
var scopePool = sync.Pool{
	New: func() interface{} { return make(Scope) },
}

// scopeLayerPool holds ScopeLayer structs for reuse.
var scopeLayerPool = sync.Pool{
	New: func() interface{} {
		return &ScopeLayer{}
	},
}

// recycleScope clears a scope map and returns it to the pool.
func recycleScope(s Scope) {
	if s != nil {
		clear(s)
		scopePool.Put(s)
	}
}

// recycleScopeLayer checks if a ScopeLayer's reference count is zero.
// If it is, the layer and its underlying Scope map are cleaned up and
// returned to their respective object pools.
func recycleScopeLayer(layer *ScopeLayer) {
	if layer == nil || layer.ref > 0 {
		return
	}
	recycleScope(layer.Scope)
	layer.Scope = nil
	layer.next = nil
	layer.ref = 0
	scopeLayerPool.Put(layer)
}
