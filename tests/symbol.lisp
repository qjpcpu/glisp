(assert (= "symbol" (type (gensym))))
(assert (= "symbol" (type (symbol "aaa"))))
(assert (= "string" (type (string 'aaa))))

(assert (= "char" (type #\')))
(assert (= "char" (type #\n)))
(assert (= "char" (type #n)))
