package extensions

import (
	"errors"
	"fmt"
	"strings"

	"github.com/qjpcpu/glisp"
)

type SexpRecordClass interface {
	glisp.Sexp
	glisp.ITypeName
	isRecordClass()
}

/* struct class */
type sexpRecordClass struct {
	typeName   string
	fieldsMeta map[string]sexpRecordField
	fieldNames []string
}

func (class *sexpRecordClass) SexpString() string {
	sb := &strings.Builder{}
	sb.WriteString("#class." + class.typeName + "\n")
	var maxLen int
	for _, name := range class.fieldNames {
		if len(name) > maxLen {
			maxLen = len(name)
		}
	}
	for _, name := range class.fieldNames {
		sb.WriteString(paddingRight(name, maxLen) + ":  " + class.fieldsMeta[name].Type + "\n")
	}
	return sb.String()
}

func (class *sexpRecordClass) isRecordClass() {}

func (class *sexpRecordClass) TypeName() string {
	return "#class." + class.typeName
}

func (class *sexpRecordClass) constructorName() string {
	return "->" + class.typeName
}

func (class *sexpRecordClass) getConstructor() *glisp.SexpFunction {
	/* build doc */
	docBuf := &strings.Builder{}
	docBuf.WriteString("Usage: (->" + class.typeName)
	var i int
	for _, f := range class.fieldsMeta {
		i++
		docBuf.WriteString(fmt.Sprintf(" '%s value%d", f.Name, i))
	}
	docBuf.WriteString(")")
	docBuf.WriteString(fmt.Sprintf("\nCreate record of type %s.", class.typeName))

	name := class.constructorName()
	return glisp.MakeUserFunction(
		name,
		buildRecordConstructor(name, class),
		glisp.WithDoc(docBuf.String()),
	)
}

func buildRecordConstructor(name string, class *sexpRecordClass) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args)%2 != 0 {
			return glisp.SexpNull, fmt.Errorf("argument of %s count must be even but got %v", name, len(args))
		}
		value, _ := glisp.MakeHash(nil)
		/* set default value */
		for _, name := range class.fieldNames {
			value.HashSet(glisp.SexpStr(name), glisp.SexpNull)
		}
		for i := 0; i < len(args); i += 2 {
			if !glisp.IsSymbol(args[i]) {
				return glisp.SexpNull, fmt.Errorf("field name must be symbol but got %s", glisp.InspectType(args[i]))
			}
			f := args[i].(glisp.SexpSymbol).Name()
			if ft, ok := class.fieldsMeta[f]; !ok {
				return glisp.SexpNull, fmt.Errorf("type %s not contains a field %s", class.typeName, f)
			} else if fv := args[i+1]; fv != glisp.SexpNull {
				if err := checkTypeMatched(ft.Type, fv); err != nil {
					return glisp.SexpNull, fmt.Errorf("field `%s` %v", ft.Name, err)
				}
			}
			value.HashSet(glisp.SexpStr(f), args[i+1])
		}
		return &SexpRecord{
			class: class,
			value: value,
		}, nil
	}
}

/* struct instance */
type SexpRecord struct {
	class *sexpRecordClass
	value *glisp.SexpHash
}

type sexpRecordField struct {
	Name string
	Type string
}

func (r *SexpRecord) SexpString() string {
	var buf strings.Builder
	buf.WriteString("#" + r.TypeName() + "{")
	var i int
	for _, name := range r.class.fieldNames {
		if i > 0 {
			buf.WriteString(", ")
		}
		v, _ := r.value.HashGet(glisp.SexpStr(name))
		buf.WriteString(name + ":" + v.SexpString())
		i++
	}
	buf.WriteString("}")
	return buf.String()
}

func (r *SexpRecord) TypeName() string {
	return r.class.typeName
}

func (t *SexpRecord) MarshalJSON() ([]byte, error) {
	return t.value.MarshalJSON()
}

func (t *SexpRecord) Explain(env *glisp.Environment, sym string, args []glisp.Sexp) (glisp.Sexp, error) {
	switch len(args) {
	case 0:
		return t.GetField(sym)
	case 1:
		return t.GetFieldDefault(sym, args[0]), nil
	default:
		return glisp.SexpNull, fmt.Errorf("record field accessor need not more than one argument but got %v", len(args))
	}
}

func (t *SexpRecord) GetField(name string) (glisp.Sexp, error) {
	if _, ok := t.class.fieldsMeta[name]; !ok {
		return glisp.SexpNull, fmt.Errorf("record<%s> not have a field named %s", t.TypeName(), name)
	}
	fv, _ := t.value.HashGet(glisp.SexpStr(name))
	return fv, nil
}

func (t *SexpRecord) GetFieldDefault(name string, defaultVal glisp.Sexp) glisp.Sexp {
	ret, err := t.GetField(name)
	if err != nil || ret == glisp.SexpNull {
		return defaultVal
	}
	return ret
}

func (t *SexpRecord) SetField(name string, val glisp.Sexp) error {
	if f, ok := t.class.fieldsMeta[name]; ok {
		if err := checkTypeMatched(f.Type, val); err != nil {
			return err
		}
		return t.value.HashSet(glisp.SexpStr(f.Name), val)
	}
	return fmt.Errorf("field %s not found", name)
}

func IsRecord(args glisp.Sexp) bool {
	_, ok := args.(*SexpRecord)
	return ok
}

func IsRecordClass(args glisp.Sexp) bool {
	_, ok := args.(SexpRecordClass)
	return ok
}

func isHashType(typ string) bool {
	return strings.HasPrefix(typ, "hash<") && strings.HasSuffix(typ, ">")
}

func isListType(typ string) bool {
	return strings.HasPrefix(typ, "list<") && strings.HasSuffix(typ, ">")
}

func getInnerType(typ string) string {
	i := strings.Index(typ, "<")
	j := strings.LastIndex(typ, ">")
	return typ[i+1 : j]
}

func getInnerKVType(typ string) (string, string) {
	var par, i int
	var k, v string
	typ = getInnerType(typ)
	for ; i < len(typ); i++ {
		if typ[i] == '<' {
			par++
		} else if typ[i] == '>' {
			par--
		}
		if typ[i] == ',' && par == 0 {
			k = typ[:i]
			v = typ[i+1:]
			break
		}
	}
	return k, v
}

func checkTypeMatched(typ string, v glisp.Sexp) error {
	if v == glisp.SexpNull {
		return nil
	}
	switch {
	case isHashType(typ):
		if !glisp.IsHash(v) {
			return fmt.Errorf("expect %s but got %v", typ, glisp.InspectType(v))
		}
		hash := v.(*glisp.SexpHash)
		ik, iv := getInnerKVType(typ)
		for _, key := range hash.KeyOrder {
			if err := checkTypeMatched(ik, key); err != nil {
				return err
			}
			val, _ := hash.HashGet(key)
			if err := checkTypeMatched(iv, val); err != nil {
				return err
			}
		}
	case isListType(typ):
		if !glisp.IsList(v) {
			return fmt.Errorf("expect %s but got %v", typ, glisp.InspectType(v))
		}
		pair := v.(*glisp.SexpPair)
		var err error
		inner := getInnerType(typ)
		pair.Foreach(func(elem glisp.Sexp) bool {
			err = checkTypeMatched(inner, elem)
			return err == nil
		})
		return err
	default:
		if glisp.InspectType(v) != typ {
			return fmt.Errorf("expect %s but got %v", typ, glisp.InspectType(v))
		}
	}
	return nil
}

func AssocRecordField(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 3 {
			return glisp.WrongNumberArguments(name, len(args), 3)
		}
		if !IsRecord(args[0]) {
			return glisp.SexpNull, fmt.Errorf("first argument must be record but got %v", glisp.InspectType(args[0]))
		}
		if !glisp.IsSymbol(args[1]) {
			return glisp.SexpNull, fmt.Errorf("second argument must be symbol but got %v", glisp.InspectType(args[1]))
		}
		record, field := args[0].(*SexpRecord), args[1].(glisp.SexpSymbol)
		return record, record.SetField(field.Name(), args[2])
	}
}

func CheckIsRecord(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 1 {
			return glisp.WrongNumberArguments(name, len(args), 1)
		}
		return glisp.SexpBool(IsRecord(args[0])), nil
	}
}

func CheckIsRecordOf(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 2 {
			return glisp.WrongNumberArguments(name, len(args), 2)
		}
		if !IsRecord(args[0]) {
			return glisp.SexpNull, fmt.Errorf("first argument must be record but got %s", glisp.InspectType(args[0]))
		}
		if !IsRecordClass(args[1]) {
			return glisp.SexpNull, fmt.Errorf("second argument must be record class but got %s", glisp.InspectType(args[1]))
		}
		cls := args[1].(SexpRecordClass)
		return glisp.SexpBool(IsRecordOf(args[0], cls.TypeName())), nil
	}
}

func IsRecordOf(r glisp.Sexp, typ string) bool {
	return IsRecord(r) && r.(*SexpRecord).class.TypeName() == typ
}

/* (defrecord MyType  (name type) (name2 type2) ) */
func DefineRecord(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) == 0 {
			return glisp.WrongNumberArguments(name, len(args), 1, glisp.Many)
		}
		if !glisp.IsSymbol(args[0]) {
			return glisp.SexpNull, fmt.Errorf("first argument must be symbol but got %v", glisp.InspectType(args[0]))
		}
		typeName := args[0].(glisp.SexpSymbol).Name()
		class := &sexpRecordClass{
			typeName:   typeName,
			fieldsMeta: make(map[string]sexpRecordField),
		}
		for _, field := range args[1:] {
			if !glisp.IsList(field) {
				return glisp.SexpNull, fmt.Errorf("field definition should be list but got %s", glisp.InspectType(field))
			}
			arr, _ := glisp.ListToArray(field)
			if len(arr) != 2 {
				return glisp.SexpNull, errors.New("field definition format must be (name type)")
			}
			if !glisp.IsSymbol(arr[0]) || !glisp.IsSymbol(arr[1]) {
				return glisp.SexpNull, errors.New("field definition format must be (name type)")
			}
			if strings.Contains(arr[1].(glisp.SexpSymbol).Name(), " ") {
				return glisp.SexpNull, errors.New("field type can't contains space")
			}
			fd := sexpRecordField{Name: arr[0].(glisp.SexpSymbol).Name(), Type: arr[1].(glisp.SexpSymbol).Name()}
			class.fieldsMeta[fd.Name] = fd
			class.fieldNames = append(class.fieldNames, fd.Name)
		}
		constructor := class.getConstructor()
		return glisp.MakeList([]glisp.Sexp{
			env.MakeSymbol("begin"),
			/* define type */
			glisp.MakeList([]glisp.Sexp{
				env.MakeSymbol("def"),
				env.MakeSymbol(class.typeName),
				class,
			}),
			/* define constructor */
			/*
			 * (defmac ->AAA [ & args ]
			 *   (let [a (->> (partition 2 (stream args))
			 *            (flatmap (fn [e] (cons (list (quote quote) (car e)) (cdr e))))
			 *            (realize))]
			 *        (syntax-quote (f (unquote-splicing a)))))
			 */
			glisp.MakeList([]glisp.Sexp{
				env.MakeSymbol("defmac"),
				env.MakeSymbol(constructor.Name()),
				glisp.SexpArray{env.MakeSymbol("&"), env.MakeSymbol("args")},
				/* let */
				glisp.MakeList([]glisp.Sexp{
					env.MakeSymbol("let"),
					glisp.SexpArray{
						env.MakeSymbol("a"),
						glisp.MakeList([]glisp.Sexp{
							env.MakeSymbol("->>"),
							glisp.MakeList([]glisp.Sexp{
								env.MakeSymbol("partition"),
								glisp.NewSexpInt(2),
								glisp.MakeList([]glisp.Sexp{
									env.MakeSymbol("stream"),
									env.MakeSymbol("args"),
								}),
							}),
							glisp.MakeList([]glisp.Sexp{
								env.MakeSymbol("flatmap"),
								glisp.MakeList([]glisp.Sexp{
									env.MakeSymbol("fn"),
									glisp.SexpArray{env.MakeSymbol("e")},
									glisp.MakeList([]glisp.Sexp{
										env.MakeSymbol("cons"),
										/* (list (quote quote) (car e)) */
										glisp.MakeList([]glisp.Sexp{
											env.MakeSymbol("list"),
											glisp.MakeList([]glisp.Sexp{
												env.MakeSymbol("quote"),
												env.MakeSymbol("quote"),
											}),
											glisp.MakeList([]glisp.Sexp{
												env.MakeSymbol("car"),
												env.MakeSymbol("e"),
											}),
										}),
										/* (cdr e) */
										glisp.MakeList([]glisp.Sexp{
											env.MakeSymbol("cdr"),
											env.MakeSymbol("e"),
										}),
									}),
								}),
							}),
							glisp.MakeList([]glisp.Sexp{env.MakeSymbol("realize")}),
						}),
					},
					glisp.MakeList([]glisp.Sexp{
						env.MakeSymbol("syntax-quote"),
						glisp.MakeList([]glisp.Sexp{
							constructor,
							glisp.MakeList([]glisp.Sexp{
								env.MakeSymbol("unquote-splicing"),
								env.MakeSymbol("a"),
							}),
						}),
					}),
				}),
			}),
		}), nil
	}
}

func paddingRight(str string, max int) string {
	if len(str) < max {
		return strings.Repeat(" ", max-len(str)) + str
	}
	return str
}

type RecordClassBuilder struct {
	cls *sexpRecordClass
}

func NewRecordClassBuilder(className string) *RecordClassBuilder {
	return &RecordClassBuilder{cls: &sexpRecordClass{typeName: className, fieldsMeta: make(map[string]sexpRecordField)}}
}

func (b *RecordClassBuilder) AddField(name string, typ string) *RecordClassBuilder {
	if _, ok := b.cls.fieldsMeta[name]; ok {
		return b
	}
	b.cls.fieldNames = append(b.cls.fieldNames, name)
	b.cls.fieldsMeta[name] = sexpRecordField{Name: name, Type: typ}
	return b
}

func (b *RecordClassBuilder) Build(env *glisp.Environment) SexpRecordClass {
	env.Bind(b.cls.typeName, b.cls)
	env.AddMacro(b.cls.constructorName(), func(_e *glisp.Environment, _args []glisp.Sexp) (glisp.Sexp, error) {
		lb := glisp.NewListBuilder()
		lb.Add(b.cls.getConstructor())
		for i := range _args {
			if i%2 == 0 {
				lb.Add(glisp.MakeList([]glisp.Sexp{
					_e.MakeSymbol("quote"),
					_args[i],
				}))
			} else {
				lb.Add(_args[i])
			}
		}
		return lb.Get(), nil
	})
	return b.cls
}