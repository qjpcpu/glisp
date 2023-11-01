package extensions

import (
	"errors"
	"fmt"
	"strings"

	"github.com/qjpcpu/glisp"
	"github.com/qjpcpu/qjson"
)

type SexpRecordClass interface {
	glisp.Sexp
	glisp.ITypeName
	MakeRecord(args []glisp.Sexp) (SexpRecord, error)
	Fields() []SexpRecordField
}

type SexpRecord interface {
	glisp.Sexp
	glisp.ITypeName
	Class() SexpRecordClass
	GetField(name string) (glisp.Sexp, error)
	GetTag(name string) glisp.SexpStr
	GetFieldDefault(name string, defaultVal glisp.Sexp) glisp.Sexp
	SetField(name string, val glisp.Sexp) error
}

/* struct class */
type sexpRecordClass struct {
	typeName   string
	fieldsMeta map[string]SexpRecordField
	fieldNames []string
}

func (class *sexpRecordClass) Cmp(o glisp.Comparable) (int, error) {
	if cls, ok := o.(SexpRecordClass); ok {
		if cls.TypeName() == class.TypeName() && len(cls.Fields()) == len(class.Fields()) {
			fs := class.Fields()
			for i, f := range cls.Fields() {
				if f.Name != fs[i].Name || f.Type != fs[i].Type || f.Tag != fs[i].Tag {
					return -1, nil
				}
			}
			return 0, nil
		}
		return -1, nil
	}
	return 0, fmt.Errorf("can't compare %v with %v", glisp.InspectType(class), glisp.InspectType(o))
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
		if tag := class.fieldsMeta[name].Tag; tag != "" {
			sb.WriteString(paddingRight(name, maxLen) + ":  " + class.fieldsMeta[name].Type + "  " + tag + "\n")
		} else {
			sb.WriteString(paddingRight(name, maxLen) + ":  " + class.fieldsMeta[name].Type + "\n")
		}
	}
	return sb.String()
}

func (class *sexpRecordClass) Fields() (fs []SexpRecordField) {
	for _, name := range class.fieldNames {
		fs = append(fs, class.fieldsMeta[name])
	}
	return
}

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
		func(_e *glisp.Environment, _a []glisp.Sexp) (glisp.Sexp, error) { return class.MakeRecord(_a) },
		glisp.WithDoc(docBuf.String()),
	)
}

func (class *sexpRecordClass) MakeRecord(args []glisp.Sexp) (SexpRecord, error) {
	if len(args)%2 != 0 {
		return nil, fmt.Errorf("argument of %s count must be even but got %v", class.constructorName(), len(args))
	}
	value, _ := glisp.MakeHash(nil)
	/* set default value */
	for _, name := range class.fieldNames {
		value.HashSet(glisp.SexpStr(name), class.fieldsMeta[name].DefaultValue)
	}
	for i := 0; i < len(args); i += 2 {
		if !glisp.IsSymbol(args[i]) {
			return nil, fmt.Errorf("field name must be symbol but got %s", glisp.InspectType(args[i]))
		}
		f := args[i].(glisp.SexpSymbol).Name()
		if ft, ok := class.fieldsMeta[f]; !ok {
			return nil, fmt.Errorf("type %s not contains a field %s", class.typeName, f)
		} else if fv := args[i+1]; fv != glisp.SexpNull {
			if err := checkTypeMatched(ft.Type, fv); err != nil {
				return nil, fmt.Errorf("field `%s` %v", ft.Name, err)
			}
		}
		value.HashSet(glisp.SexpStr(f), args[i+1])
	}
	return &sexpRecord{
		class: class,
		value: value,
	}, nil
}

/* struct instance */
type sexpRecord struct {
	class *sexpRecordClass
	value *glisp.SexpHash
}

type SexpRecordField struct {
	Name         string
	Type         string
	Tag          string
	DefaultValue glisp.Sexp
}

func (r *sexpRecord) SexpString() string {
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

func (r *sexpRecord) TypeName() string {
	return r.class.typeName
}

func (r *sexpRecord) Class() SexpRecordClass { return r.class }

func (t *sexpRecord) MarshalJSON() ([]byte, error) {
	return t.value.MarshalJSON()
}

func (t *sexpRecord) Explain(env *glisp.Environment, sym string, args []glisp.Sexp) (glisp.Sexp, error) {
	switch len(args) {
	case 0:
		if strings.HasSuffix(sym, ".tag") {
			return t.GetTag(strings.TrimSuffix(sym, ".tag")), nil
		}
		return t.GetField(sym)
	case 1:
		return t.GetFieldDefault(sym, args[0]), nil
	default:
		return glisp.SexpNull, fmt.Errorf("record field accessor need not more than one argument but got %v", len(args))
	}
}

func (t *sexpRecord) GetTag(name string) glisp.SexpStr {
	return glisp.SexpStr(t.class.fieldsMeta[name].Tag)
}

func (t *sexpRecord) GetField(name string) (glisp.Sexp, error) {
	if _, ok := t.class.fieldsMeta[name]; !ok {
		return glisp.SexpNull, fmt.Errorf("record<%s> not have a field named %s", t.TypeName(), name)
	}
	fv, _ := t.value.HashGet(glisp.SexpStr(name))
	return fv, nil
}

func (t *sexpRecord) GetFieldDefault(name string, defaultVal glisp.Sexp) glisp.Sexp {
	ret, err := t.GetField(name)
	if err != nil || ret == glisp.SexpNull {
		return defaultVal
	}
	return ret
}

func (t *sexpRecord) SetField(name string, val glisp.Sexp) error {
	if f, ok := t.class.fieldsMeta[name]; ok {
		if err := checkTypeMatched(f.Type, val); err != nil {
			return err
		}
		return t.value.HashSet(glisp.SexpStr(f.Name), val)
	}
	return fmt.Errorf("field %s not found", name)
}

func IsRecord(args glisp.Sexp) bool {
	_, ok := args.(SexpRecord)
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
	userfn := func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if !IsRecord(args[0]) {
			return glisp.SexpNull, fmt.Errorf("first argument must be record but got %v", glisp.InspectType(args[0]))
		}
		var field string
		switch expr := args[1].(type) {
		case glisp.SexpStr:
			field = string(expr)
		case glisp.SexpSymbol:
			field = expr.Name()
		default:
			return glisp.SexpNull, fmt.Errorf("second argument must be symbol/string but got %v", glisp.InspectType(args[1]))
		}
		record := args[0].(SexpRecord)
		return record, record.SetField(field, args[2])
	}
	sexpfn := glisp.MakeUserFunction(name, userfn)
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 3 {
			return glisp.WrongNumberArguments(name, len(args), 3)
		}
		key := args[1]
		if glisp.IsSymbol(args[1]) {
			key = glisp.MakeList([]glisp.Sexp{env.MakeSymbol("quote"), args[1]})
		}
		return glisp.MakeList([]glisp.Sexp{
			sexpfn,
			args[0],
			key,
			args[2],
		}), nil
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

func GetRecordClass(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 1 {
			return glisp.WrongNumberArguments(name, len(args), 1)
		}
		if !IsRecord(args[0]) {
			return glisp.SexpNull, fmt.Errorf("%v is not record", glisp.InspectType(args[0]))
		}
		return args[0].(SexpRecord).Class(), nil
	}
}

func CheckIsRecordClass(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 1 {
			return glisp.WrongNumberArguments(name, len(args), 1)
		}
		return glisp.SexpBool(IsRecordClass(args[0])), nil
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

func ClassDefinition(name string) glisp.UserFunction {
	return func(env *glisp.Environment, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 1 {
			return glisp.WrongNumberArguments(name, len(args), 1)
		}
		if !IsRecordClass(args[0]) {
			return glisp.SexpNull, fmt.Errorf("first argument must be record class but got %s", glisp.InspectType(args[0]))
		}
		cls := args[0].(SexpRecordClass)
		var fields glisp.SexpArray
		for _, f := range cls.Fields() {
			hash, _ := glisp.MakeHash([]glisp.Sexp{
				glisp.SexpStr("name"),
				glisp.SexpStr(f.Name),
				glisp.SexpStr("type"),
				glisp.SexpStr(f.Type),
				glisp.SexpStr("tag"),
				glisp.SexpStr(f.Tag),
			})
			fields = append(fields, hash)
		}
		return glisp.MakeHash([]glisp.Sexp{
			glisp.SexpStr("name"),
			glisp.SexpStr(strings.TrimPrefix(cls.TypeName(), "#class.")),
			glisp.SexpStr("fields"),
			fields,
		})
	}
}

func IsRecordOf(r glisp.Sexp, typ string) bool {
	return IsRecord(r) && r.(SexpRecord).Class().TypeName() == typ
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
			fieldsMeta: make(map[string]SexpRecordField),
		}
		for _, field := range args[1:] {
			if !glisp.IsList(field) {
				return glisp.SexpNull, fmt.Errorf("field definition should be list but got %s", glisp.InspectType(field))
			}
			arr, _ := glisp.ListToArray(field)
			if argNum := len(arr); argNum < 2 || argNum > 4 {
				return glisp.SexpNull, errors.New("field definition format must be (name type) or (name type tag) or (name type tag default-value)")
			}
			if !glisp.IsSymbol(arr[0]) || !glisp.IsSymbol(arr[1]) {
				return glisp.SexpNull, errors.New("field definition format must be (name type) or (name type tag)")
			}
			if strings.Contains(arr[1].(glisp.SexpSymbol).Name(), " ") {
				return glisp.SexpNull, errors.New("field type can't contains space")
			}
			if (len(arr) == 3 || len(arr) == 4) && !glisp.IsString(arr[2]) {
				if err := env.LoadExpressions([]glisp.Sexp{arr[2]}); err != nil {
					return glisp.SexpNull, fmt.Errorf("eval field %s tag %s fail %v", arr[0].SexpString(), arr[2].SexpString(), err)
				}
				if expr, err := env.Run(); err != nil {
					return glisp.SexpNull, fmt.Errorf("eval field %s tag %s fail %v", arr[0].SexpString(), arr[2].SexpString(), err)
				} else if !glisp.IsString(expr) {
					return glisp.SexpNull, errors.New("field definition format must be (name type) or (name type tag) or (name type tag default-value)")
				} else {
					arr[2] = expr
				}
			}
			fd := SexpRecordField{Name: arr[0].(glisp.SexpSymbol).Name(), Type: arr[1].(glisp.SexpSymbol).Name(), DefaultValue: glisp.SexpNull}
			if len(arr) >= 3 {
				fd.Tag = string(arr[2].(glisp.SexpStr))
			}
			/* default value */
			if len(arr) == 4 {
				if err := env.LoadExpressions([]glisp.Sexp{arr[3]}); err != nil {
					return glisp.SexpNull, err
				}
				if dfv, err := env.Run(); err != nil {
					return glisp.SexpNull, err
				} else {
					fd.DefaultValue = dfv
				}
			}
			class.fieldsMeta[fd.Name] = fd
			class.fieldNames = append(class.fieldNames, fd.Name)
		}
		constructor := class.getConstructor()
		var_args, var_a, var_e := env.GenSymbol(), env.GenSymbol(), env.GenSymbol()
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
				glisp.SexpArray{env.MakeSymbol("&"), var_args},
				/* let */
				glisp.MakeList([]glisp.Sexp{
					env.MakeSymbol("let"),
					glisp.SexpArray{
						var_a,
						glisp.MakeList([]glisp.Sexp{
							env.MakeSymbol("->>"),
							glisp.MakeList([]glisp.Sexp{
								env.MakeSymbol("partition"),
								glisp.NewSexpInt(2),
								glisp.MakeList([]glisp.Sexp{
									env.MakeSymbol("stream"),
									var_args,
								}),
							}),
							glisp.MakeList([]glisp.Sexp{
								env.MakeSymbol("flatmap"),
								glisp.MakeList([]glisp.Sexp{
									env.MakeSymbol("fn"),
									glisp.SexpArray{var_e},
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
												var_e,
											}),
										}),
										/* (cdr e) */
										glisp.MakeList([]glisp.Sexp{
											env.MakeSymbol("cdr"),
											var_e,
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
								var_a,
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
	return &RecordClassBuilder{cls: &sexpRecordClass{typeName: className, fieldsMeta: make(map[string]SexpRecordField)}}
}

func (b *RecordClassBuilder) AddField(name string, typ string) *RecordClassBuilder {
	return b.AddFullField(name, typ, "", glisp.SexpNull)
}

func (b *RecordClassBuilder) AddFullField(name, typ, tag string, defaultValue glisp.Sexp) *RecordClassBuilder {
	if _, ok := b.cls.fieldsMeta[name]; ok {
		return b
	}
	b.cls.fieldNames = append(b.cls.fieldNames, name)
	b.cls.fieldsMeta[name] = SexpRecordField{Name: name, Type: typ, DefaultValue: defaultValue, Tag: tag}
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

/* record accessor */
func ToGoRecord(r SexpRecord) *SexpGoRecord { return &SexpGoRecord{SexpRecord: r} }

type SexpGoRecord struct {
	SexpRecord
}

func (r *SexpGoRecord) GetTag(name string) string {
	return string(r.SexpRecord.GetTag(name))
}

func (r *SexpGoRecord) GetStringField(name string) string {
	return string(r.SexpRecord.GetFieldDefault(name, glisp.SexpStr("")).(glisp.SexpStr))
}

func (r *SexpGoRecord) GetBoolField(name string) bool {
	return bool(r.SexpRecord.GetFieldDefault(name, glisp.SexpBool(false)).(glisp.SexpBool))
}

func (r *SexpGoRecord) GetIntField(name string) int64 {
	return r.SexpRecord.GetFieldDefault(name, glisp.NewSexpInt64(0)).(glisp.SexpInt).ToInt64()
}

func (r *SexpGoRecord) GetUintField(name string) uint64 {
	return r.SexpRecord.GetFieldDefault(name, glisp.NewSexpInt64(0)).(glisp.SexpInt).ToUint64()
}

func (r *SexpGoRecord) GetBytesField(name string) []byte {
	return r.SexpRecord.GetFieldDefault(name, glisp.NewSexpBytes(nil)).(glisp.SexpBytes).Bytes()
}

func (r *SexpGoRecord) GetHashField(name string) map[string]interface{} {
	ret := make(map[string]interface{})
	h, _ := glisp.MakeHash(nil)
	bs, _ := glisp.Marshal(r.SexpRecord.GetFieldDefault(name, h))
	stdUnmarshal(bs, &ret)
	return ret
}

func (r *SexpGoRecord) GetListField(name string) (ret []interface{}) {
	bs, _ := glisp.Marshal(r.SexpRecord.GetFieldDefault(name, glisp.SexpArray{}))
	stdUnmarshal(bs, &ret)
	return ret
}

func (r *SexpGoRecord) SetStringField(name string, val string) *SexpGoRecord {
	r.SexpRecord.SetField(name, glisp.SexpStr(val))
	return r
}

func (r *SexpGoRecord) SetBoolField(name string, val bool) *SexpGoRecord {
	r.SexpRecord.SetField(name, glisp.SexpBool(val))
	return r
}

func (r *SexpGoRecord) SetIntField(name string, val int64) *SexpGoRecord {
	r.SexpRecord.SetField(name, glisp.NewSexpInt64(val))
	return r
}

func (r *SexpGoRecord) SetUintField(name string, val uint64) *SexpGoRecord {
	r.SexpRecord.SetField(name, glisp.NewSexpUint64(val))
	return r
}

func (r *SexpGoRecord) SetBytesField(name string, val []byte) *SexpGoRecord {
	r.SexpRecord.SetField(name, glisp.NewSexpBytes(val))
	return r
}

func (r *SexpGoRecord) SetHashField(name string, val map[string]interface{}) *SexpGoRecord {
	bytes := qjson.JSONMarshalWithPanic(val)
	hash, _ := ParseJSON(bytes)
	r.SexpRecord.SetField(name, hash)
	return r
}

func (r *SexpGoRecord) SetListField(name string, val []interface{}) *SexpGoRecord {
	bytes := qjson.JSONMarshalWithPanic(val)
	list, _ := ParseJSON(bytes)
	if glisp.IsArray(list) {
		for _, f := range r.Class().Fields() {
			if f.Name == name {
				if strings.HasPrefix(f.Type, "list") {
					arr := []glisp.Sexp(list.(glisp.SexpArray))
					r.SexpRecord.SetField(name,
						glisp.NewListBuilder().Add(arr...).Get())
				} else {
					r.SexpRecord.SetField(name, list)
				}
				return r
			}
		}
	}
	r.SexpRecord.SetField(name, list)
	return r
}
