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
(assert (= "{\"a\":1,\"b\":null,\"c\":true,\"hash\":{\"a\":1,\"b\":null},\"list\":[\"element\",null],\"array\":[1,2,3]}" (json/stringify h1)))


(def h2 {'a (time/parse 1656988237)})
(assert (= "{\"a\":\"2022-07-05 10:30:37\"}" (json/stringify h2)))

(def js (json/parse (os/read-file "./test-data.json")))
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

(assert (= "\"fn:+\"" (json/stringify +)))
(assert (= "97" (json/stringify #a)))

(assert (= "[1,2,3]" (json/stringify '(1 2 3))))

;; parse empty string
(assert (= '() (json/parse "" '())))
(assert (= '() (json/parse "  " '())))
(assert (= '() (json/parse " \t " '())))
(assert (= '() (json/parse (bytes "") '())))
(assert (= '() (json/parse "null")))
(assert (= '() (json/parse (bytes "null"))))

;; parse json atom
(assert (= 1 (json/parse 1)))
(assert (= 1.0 (json/parse 1.0)))
(assert (= true (json/parse true)))

;; with indent
(assert (= "{
  \"a\": 1
}" (json/stringify {"a" 1} true)))


(assert (= 1 (json/query {"a" 1} "b" 1)))
(assert (= 1 (json/query '() "b" 1)))

(assert (= [1] (json/parse "[1" [1])))
(assert (= "hash" (type (json/parse "[1" {}))))

(assert (= 1 (json/parse '() 1)))

;; json marshal float shouldn't lost precision
(assert (= "{\"a\":119.38907218102548}" (json/stringify {"a" (+ 1 118.38907218102548)})))
(assert (= "119.38907218102548" (string (+ 1 118.38907218102548))))
(assert (= "119.38907218102548" (sexp-str (+ 1 118.38907218102548))))

(assert (= "[\"nil\",1,2,\"nil\"]" (json/stringify '(nil 1 2 nil))))
(assert (= "[null,1,2,null]" (json/stringify (list nil 1 2 nil))))
(assert (= "[1,2,null]" (json/stringify '(1 2 ()))))
(assert (= "[1,2]" (json/stringify (cons 1 2 ))))

(def complex-json {"Result" {"Tasks" [{"ProgressSummary" {"Infos" [{"NameCN" "cn1" "ProgressType" "Incr" "Status" "Todo"}]} "TrafficSpec" "Standard"} {"ProgressSummary" {"Infos" [{"NameCN" "cn2" "ProgressType" "Incr" "Status" "Todo"}]} "TrafficSpec" "Standard"}] "Total" 1162}})
(assert (= ["cn1" "cn2"]
           (json/query complex-json "Result.Tasks.#.ProgressSummary.Infos.#(ProgressType==Incr).NameCN")))
(json/set complex-json "Result.Tasks.#.ProgressSummary.Infos.#(ProgressType==Incr).NameCN" "CN")
(assert (= ["CN" "CN"]
           (json/query complex-json "Result.Tasks.#.ProgressSummary.Infos.#(ProgressType==Incr).NameCN")))

(def complex-json {"Result" {"Tasks" [{"Number" 1 "ProgressSummary" {"Infos" [{"NameCN" "cn1" "ProgressType" "Incr" "Status" "Todo"}]} "TrafficSpec" "Standard"} {"ProgressSummary" {"Infos" [{"NameCN" "cn2" "ProgressType" "Incr" "Status" "Todo"}]} "TrafficSpec" "Standard"}] "Total" 1162}})
(assert (= "cn1"
           (json/query complex-json "Result.Tasks.#(Number==1).ProgressSummary.Infos.#(ProgressType==Incr).NameCN")))
(json/set complex-json "Result.Tasks.#(Number==1).ProgressSummary.Infos.#(ProgressType==Incr).NameCN" "CN")
(assert (= ["CN" "cn2"]
           (json/query complex-json "Result.Tasks.#.ProgressSummary.Infos.#(ProgressType==Incr).NameCN")))

(def complex-json {"Result" {"Tasks" [{"ProgressSummary" {"Infos" [{"NameCN" "cn1" "ProgressType" "Incr" "Status" "Todo"}]} "TrafficSpec" "Standard"} {"ProgressSummary" {"Infos" [{"NameCN" "cn2" "ProgressType" "Incr" "Status" "Todo"}]} "TrafficSpec" "Standard"}] "Total" 1162}})
(json/del complex-json "Result.Tasks.#")
(assert (= "{\"Result\":{\"Tasks\":[],\"Total\":1162}}" (json/stringify complex-json)))

(def complex-json {"Result" {"Tasks" [{"ProgressSummary" {"Infos" [{"NameCN" "cn1" "ProgressType" "Incr" "Status" "Todo"}]} "TrafficSpec" "Standard"} {"ProgressSummary" {"Infos" [{"NameCN" "cn2" "ProgressType" "Incr" "Status" "Todo"}]} "TrafficSpec" "Standard"}] "Total" 1162}})
(json/del complex-json "Result.Tasks.#.ProgressSummary.Infos.#(ProgressType==Incr).NameCN")
(assert (= "{\"Result\":{\"Tasks\":[{\"ProgressSummary\":{\"Infos\":[{\"ProgressType\":\"Incr\",\"Status\":\"Todo\"}]},\"TrafficSpec\":\"Standard\"},{\"ProgressSummary\":{\"Infos\":[{\"ProgressType\":\"Incr\",\"Status\":\"Todo\"}]},\"TrafficSpec\":\"Standard\"}],\"Total\":1162}}"
           (json/stringify complex-json)))

(defn hash-eq [j1 j2] (assert (= (json/stringify j1) (json/stringify j2))))
(def js (json/sadd (json/sadd {"a" 1} "b" 1) "b" 1))
(hash-eq {"a" 1 "b" [1]} js)

(def js (json/add (json/add {"a" 1} "b" 1) "b" 1))
(hash-eq {"a" 1 "b" [1 1]} js)
