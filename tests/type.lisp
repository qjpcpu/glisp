(assert (= (typestr 'a) "symbol"))
(assert (= (typestr "a") "string"))
(assert (= (typestr [1]) "array"))
(assert (= (typestr true) "bool"))
(assert (= (typestr #a) "char"))
(assert (= (typestr 1.1) "float"))
(assert (= (typestr (fn [a] ())) "function"))
(assert (= (typestr (fn [a] ())) "function"))
(assert (= (typestr {'a 1} ) "hash"))
(assert (= (typestr 1) "int"))
(assert (= (typestr '('a)) "list"))
(assert (= (typestr '()) "list"))
(assert (= (typestr (time/now)) "time"))
(assert (= (typestr (make-chan)) "channel"))
(assert (= (typestr 0B4c6561726e20476f21) "bytes"))
(assert (= (string 0B676c69737020697320636f6f6c) "glisp is cool"))
(assert (= (bytes "glisp is cool") 0B676c69737020697320636f6f6c))
(assert (= 0B676c69737020697320636f6f6c (bytes "glisp is cool") ))
(assert (bytes? 0B6869))
(assert (bool? false))
(assert (= "symbol" (typestr (symbol 'a))))
