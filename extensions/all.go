package extensions

import "github.com/qjpcpu/glisp"

func ImportAll(env *glisp.Environment) error {
	modules := []func(*glisp.Environment) error{
		func(e *glisp.Environment) error { return e.ImportEval() },
		ImportCoreUtils,
		ImportRandom,
		ImportMathUtils,
		ImportTime,
		ImportChannels,
		ImportCoroutines,
		ImportRegex,
		ImportBase64,
		ImportJSON,
		ImportString,
		ImportOS,
		ImportHTTP,
		ImportCSV,
	}
	for _, f := range modules {
		if err := f(env); err != nil {
			return err
		}
	}
	return nil
}
