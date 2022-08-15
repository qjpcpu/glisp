(assert (null? (query '() "")))
;; bad selector
(assert (null? (query {} "#((")))

(def h {'a 1 "b" "hello" 1 100 "null" '() "arr" [1 2 3]})
(assert (= 1 (query h "a")))
(assert (= "hello" (query h "b")))
(assert (= 100 (query h "1")))
(assert (= '() (query h "null")))
(assert (= '() (query h "null.e")))
(assert (= 2 (query h "arr.1")))
(assert (= '() (query h "arr.100")))
(assert (= [3] (query h "arr.#(>2)")))
(assert (= '() (query h "b.t.v")))

(def h1 {"a1" [{"name" "jack" "age" 20 "height" 174.3} {"name" "tom" "age" 30 "height" 160.3}]})
(assert (= ["tom"] (query h1 "a1.#(age>20).name")))
(assert (= ["jack" "tom"] (query h1 "a1.#(age>=20).name")))
(assert (= ["jack"] (query h1 "a1.#(age<30).name")))
(assert (= ["jack" "tom"] (query h1 "a1.#(age<=30).name")))

(assert (= ["jack"] (query h1 "a1.#(height>170).name")))
(assert (= ["jack" "tom"] (query h1 "a1.#(height>=160.3).name")))
(assert (= ["tom"] (query h1 "a1.#(height<174).name")))
(assert (= ["tom"] (query h1 "a1.#(height<=160.3).name")))

(assert (= ["jack"] (query h1 "a1.#(age==20).name")))
(assert (= ["jack"] (query h1 "a1.#(age=2).name")))
(assert (= ["tom"] (query h1 "a1.#(age!=20).name")))
(assert (= ["tom"] (query h1 "a1.#(age!==20).name")))

(def arr [{"age" 20 "name" "jack"} {"age" 30 "name" "tom"}])
(assert (= ["tom"] (query arr "#(age>20).name")))

;; escape
(def h2 {"fav.movie" "Deer Hunter"})
(assert (= "Deer Hunter" (query h2 "fav\\.movie")))

(def h3 {"a1" [{"name" "jack" "age" "20" "height" 174.3} {"name" "tom" "age" "30" "height" 160.3}]})
(assert (= ["tom"] (query h3 "a1.#(age>20).name")))
(assert (= ["jack" "tom"] (query h3 "a1.#(age>=20).name")))
(assert (= ["jack"] (query h3 "a1.#(age<30).name")))
(assert (= ["jack" "tom"] (query h3 "a1.#(age<=30).name")))
