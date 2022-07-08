(assert (= "symbol" (typestr (gensym))))
(assert (= "symbol" (typestr (str2sym "aaa"))))
(assert (= "string" (typestr (sym2str 'aaa))))

(assert (= "char" (typestr #\')))
(assert (= "char" (typestr #\n)))
(assert (= "char" (typestr #n)))
