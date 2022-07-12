(def h1 {'a 1 'b '() 'c true})
(assert (= "{\"a\":1,\"b\":null,\"c\":true}" (json/stringify h1)))

(def a1 ['a 1 '() false "str"])
(assert (= "[\"a\",1,null,false,\"str\"]" (json/stringify a1)))


(def c1 {'a 1 'b '()})
(def c2 '("element" ()))
(def c3 [1 2 3])
(hset! h1 'hash c1)
(hset! h1 'list c2)
(hset! h1 'array c3)
(assert (= "{\"a\":1,\"b\":null,\"c\":true,\"hash\":{\"a\":1,\"b\":null},\"list\":[\"element\"],\"array\":[1,2,3]}" (json/stringify h1)))


(def h2 {'a (time/parse 1656988237)})
(assert (= "{\"a\":\"2022-07-05 10:30:37\"}" (json/stringify h2)))

(def js (json/parse (test/read-file "./test-data.json")))
(assert (> (len (json/stringify js)) 0))
(assert (= 100 (hget js "number")))
(assert (= 1.23 (hget js "float")))
(assert (= "hello" (hget js "string")))
(assert (hget (hget js "hash") "a"))
(assert (not (hget (hget js "hash") "b")))
(assert (= '() (hget js "nothing")))
(assert (= 5 (len (hget js "list"))))

(def item (aget (hget js "list") 0))
(assert (= "5e7dbc9fd0cc8370c563a1d7" (hget item "_id")))
(assert (hget item "isActive"))
(assert (=  "\"Fuller\"" (json/stringify (hget (hget item "name") "first"))))


;; json with bytes
(def jb {'a 0B676c69737020697320636f6f6c})
(assert (= "{\"a\":\"Z2xpc3AgaXMgY29vbA==\"}" (json/stringify jb)))
(assert (= 0B676c69737020697320636f6f6c  (base64/decode "Z2xpc3AgaXMgY29vbA==")))
(assert (= (base64/encode 0B676c69737020697320636f6f6c)  "Z2xpc3AgaXMgY29vbA=="))

(assert (= [1] (json/parse "[1]")))
