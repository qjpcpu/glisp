package glisp

// Opcode is the type for virtual machine instructions.
type Opcode byte

const (
	// OpInvalid is an invalid instruction
	OpInvalid Opcode = iota

	// Stack manipulation
	OpPush
	OpPushClosure
	OpPop
	OpDup

	// Variable access
	OpGet
	OpPut
	OpBindDynFun

	// Control flow
	OpJump      // Unconditional relative jump
	OpGoto      // Unconditional absolute jump
	OpBranch    // Conditional relative jump
	OpReturn    // Return from function
	OpCall      // Call a function by symbol
	OpPrepare   // Prepare for tail call
	OpDispatch  // Call a function from stack
	OpUserInstr // Call a user-defined instruction

	// Scope
	OpAddScope
	OpRemoveScope

	// Special instructions
	OpExplode
	OpSquash
	OpVectorize
	OpHashize
	OpBindlist
	OpRefSym
)
