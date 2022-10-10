(assert (nil? (json/query '() "")))
;; bad selector
(assert (nil? (json/query {} "#((")))

(def h {'a 1 "b" "hello" 1 100 "null" '() "arr" [1 2 3]})
(assert (= 1 (json/query h "a")))
(assert (= "hello" (json/query h "b")))
(assert (= 100 (json/query h "1")))
(assert (= '() (json/query h "null")))
(assert (= '() (json/query h "null.e")))
(assert (= 2 (json/query h "arr.1")))
(assert (= '() (json/query h "arr.100")))
(assert (= 3 (json/query h "arr.#(>2)")))
(assert (= '() (json/query h "b.t.v")))

(def h1 {"a1" [{"name" "jack" "age" 20 "height" 174.3} {"name" "tom" "age" 30 "height" 160.3}]})
(assert (= "tom" (json/query h1 "a1.#(age>20).name")))
(assert (= nil (json/query h1 "a1.#(age>200).name")))
(assert (= ["jack" "tom"] (json/query h1 "a1.#(age>=20).name")))
(assert (= "jack" (json/query h1 "a1.#(age<30).name")))
(assert (= ["jack" "tom"] (json/query h1 "a1.#(age<=30).name")))

(assert (= "jack" (json/query h1 "a1.#(height>170).name")))
(assert (= ["jack" "tom"] (json/query h1 "a1.#(height>=160.3).name")))
(assert (= "tom" (json/query h1 "a1.#(height<174).name")))
(assert (= "tom" (json/query h1 "a1.#(height<=160.3).name")))

(assert (= "jack" (json/query h1 "a1.#(age==20).name")))
(assert (= "jack" (json/query h1 "a1.#(age=2).name")))
(assert (= "tom" (json/query h1 "a1.#(age!=20).name")))
(assert (= "tom" (json/query h1 "a1.#(age!==20).name")))

(def arr [{"age" 20 "name" "jack"} {"age" 30 "name" "tom"}])
(assert (= "tom" (json/query arr "#(age>20).name")))

;; escape
(def h2 {"fav.movie" "Deer Hunter"})
(assert (= "Deer Hunter" (json/query h2 "fav\\.movie")))

(def h3 {"a1" [{"name" "jack" "age" "20" "height" 174.3} {"name" "tom" "age" "30" "height" 160.3}]})
(assert (= "tom" (json/query h3 "a1.#(age>20).name")))
(assert (= ["jack" "tom"] (json/query h3 "a1.#(age>=20).name")))
(assert (= "jack" (json/query h3 "a1.#(age<30).name")))
(assert (= ["jack" "tom"] (json/query h3 "a1.#(age<=30).name")))

(def s1 {'a 1 "b" "hello" 1 100 "null" '() "arr" [1 2 3]})
(let [v (json/set s1 "c" 100)]
     (assert (= (hget v "c") 100)))

(def s1 {'a 1 "b" "hello" 1 100 "null" '() "arr" [1 2 3]})
(let [v (json/set s1 "d.e" 100)]
     (assert (= (json/query v "d.e") 100)))

(def s1 {'a 1 "b" "hello" 1 100 "null" '() "arr" [1 2 3]})
(let [v (json/set s1 "arr.1" 100)]
     (assert (= (hget v "arr") [1 100 3])))

(def s1 {'a 1 "b" "hello" 1 100 "null" '() "arr" [1 2 3]})
(let [v (json/set s1 "arr.5" 100)]
     (assert (= (hget v "arr") [1 2 3 100])))

(def s1 {'a 1 "b" "hello" 1 100 "null" '() "arr" [1 {} 3]})
(let [v (json/set s1 "arr.1.key" 100)]
     (assert (= (json/query v "arr.0") 1))
     (assert (= (json/query v "arr.2") 3))
     (assert (= (json/query v "arr.1.key")  100)))

(def s1 {'a 1 "b" "hello" 1 100 "null" '() "arr" [1 2 3]})
(let [v (json/set s1 "arr.10.key" 100)]
     (assert (= (json/query v "arr.0") 1))
     (assert (= (json/query v "arr.1") 2))
     (assert (= (json/query v "arr.2") 3))
     (assert (= (json/query v "arr.3.key")  100)))

(def s1 {'a 1 "b" "hello" 1 100 "null" '() "arr" [1 2 3]})
(let [v (json/del s1 "b")]
     (assert (not (exist? v "b"))))

(def s1 {'a 1 "b" "hello" 1 100 "obj" {"inner" 1} "arr" [1 2 3]})
(let [v (json/del s1 "obj.inner")]
     (assert (empty? (hget v "obj"))))

(def s1 {'a 1 "b" "hello" 1 100 "obj" {"inner" 1} "arr" [1 2 3]})
(let [v (json/del s1 "obj")]
     (assert (not (exist? v "obj"))))

(def s1 {'a 1 "b" "hello" 1 100 "obj" {"inner" 1} "arr" [1 2 3]})
(let [v (json/del s1 "arr.1")]
     (assert (= [1 3] (hget v "arr"))))

(def s1 [{'a 1 "b" "hello" 1 100 "obj" {"inner" 1} "arr" [1 2 3]}])
(let [v (json/del s1 "0.arr.1")]
     (assert (= [1 3] (json/query v "0.arr"))))

(def s1 [{'a 1 "b" "hello" 1 100 "obj" {"inner" 1} "arr" [1 2 3]}])
(let [v (json/del s1 "0.arr.100")]
     (assert (= [1 2 3] (json/query v "0.arr"))))

(def s1 [{'a 1 "b" "hello" 1 100 "obj" {"inner" 1} "arr" [1 2 3]}])
(let [v (json/del s1 "1000.arr.100")]
     (assert (= (json/stringify [{'a 1 "b" "hello" 1 100 "obj" {"inner" 1} "arr" [1 2 3]}]) (json/stringify v))))

(def s1 {'a 1 "b" "hello" 1 100 "obj" {"inner" 1} "arr" [1 2 3]})
(let [v (json/del s1 "brr.1")]
     (assert (= (json/stringify {'a 1 "b" "hello" 1 100 "obj" {"inner" 1} "arr" [1 2 3]}) (json/stringify v))))

(def s1 {"a" {"b" '()}})
(let [v (json/del s1 "a.b.1")]
     (assert (= (json/stringify {"a" {"b" '()}}) (json/stringify v))))

(def s1 {"a" {"b" '()}})
(let [v (json/del s1 "a.b")]
     (assert (= (json/stringify {"a" {}}) (json/stringify v))))

(def s1 {"a" {"b" '()}})
(let [v (json/del s1 "a")]
     (assert (= (json/stringify {}) (json/stringify v))))

;; set single
(def hx {"a1" [{"name" "jack" "age" 20} {"name" "tom" "age" 30}]})
(let [v (json/set hx "a1.#(age>20).name" "single")]
     (assert (= #`{"a1":[{"name":"jack","age":20},{"name":"single","age":30}]}` (json/stringify v))))
;; set multiple
(def hx {"a1" [{"name" "jack" "age" 20} {"name" "tom" "age" 30}]})
(let [v (json/set hx "a1.#(age>10).name" "same")]
     (assert (= #`{"a1":[{"name":"same","age":20},{"name":"same","age":30}]}` (json/stringify v))))
;; set nothing
(def hx {"a1" [{"name" "jack" "age" 20} {"name" "tom" "age" 30}]})
(let [v (json/set hx "a1.#(age>100).name" "same")]
     (assert (= #`{"a1":[{"name":"jack","age":20},{"name":"tom","age":30}]}` (json/stringify v))))

;; del single
(def hx {"a1" [{"name" "jack" "age" 20} {"name" "tom" "age" 30}]})
(let [v (json/del hx "a1.#(age>20).name")]
     (assert (= #`{"a1":[{"name":"jack","age":20},{"age":30}]}` (json/stringify v))))

;; del multiple
(def hx {"a1" [{"name" "jack" "age" 20} {"name" "tom" "age" 30}]})
(let [v (json/del hx "a1.#(age>10).name")]
     (assert (= #`{"a1":[{"age":20},{"age":30}]}` (json/stringify v))))

;; del nothing
(def hx {"a1" [{"name" "jack" "age" 20} {"name" "tom" "age" 30}]})
(let [v (json/del hx "a1.#(age>100).name")]
     (assert (= #`{"a1":[{"name":"jack","age":20},{"name":"tom","age":30}]}` (json/stringify v))))
