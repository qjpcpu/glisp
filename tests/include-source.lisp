
; (println "Here in base ")

(source-file "./inc.lisp" "./inc1.lisp")

; (println "Calling function defined in inc.lisp")

(assert (= (simple) "from include"))

; (println "Calling function defined in inc1.lisp")

(assert (= (simple1) "from include 1"))

; (println (simple))
; (println (simple1))

(include "./inc2.lisp" "./inc3.lisp")

(assert (= (inc2) "from inc2.lisp"))
(assert (= (inc3) "from inc3.lisp"))
