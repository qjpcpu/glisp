(assert (= "symbol" (typestr (gensym))))
(assert (= "symbol" (typestr (symbol "aaa"))))
(assert (= "string" (typestr (str 'aaa))))

(assert (= "char" (typestr #\')))
(assert (= "char" (typestr #\n)))
(assert (= "char" (typestr #n)))
