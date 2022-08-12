(assert (null? (hsearch '() "")))
;; bad selector
(assert (null? (hsearch {} "#((")))

(def h {'a 1 "b" "hello" 1 100 "null" '() "arr" [1 2 3]})
(assert (= 1 (hsearch h "a")))
(assert (= "hello" (hsearch h "b")))
(assert (= 100 (hsearch h "1")))
(assert (= '() (hsearch h "null")))
(assert (= '() (hsearch h "null.e")))
(assert (= 2 (hsearch h "arr.1")))
(assert (= '() (hsearch h "arr.100")))
(assert (= [3] (hsearch h "arr.#(>2)")))
(assert (= '() (hsearch h "b.t.v")))

(def h1 {"a1" [{"name" "jack" "age" 20 "height" 174.3} {"name" "tom" "age" 30 "height" 160.3}]})
(assert (= ["tom"] (hsearch h1 "a1.#(age>20).name")))
(assert (= ["jack" "tom"] (hsearch h1 "a1.#(age>=20).name")))
(assert (= ["jack"] (hsearch h1 "a1.#(age<30).name")))
(assert (= ["jack" "tom"] (hsearch h1 "a1.#(age<=30).name")))

(assert (= ["jack"] (hsearch h1 "a1.#(height>170).name")))
(assert (= ["jack" "tom"] (hsearch h1 "a1.#(height>=160.3).name")))
(assert (= ["tom"] (hsearch h1 "a1.#(height<174).name")))
(assert (= ["tom"] (hsearch h1 "a1.#(height<=160.3).name")))

(assert (= ["jack"] (hsearch h1 "a1.#(age==20).name")))
(assert (= ["jack"] (hsearch h1 "a1.#(age=2).name")))
(assert (= ["tom"] (hsearch h1 "a1.#(age!=20).name")))
(assert (= ["tom"] (hsearch h1 "a1.#(age!==20).name")))

(def arr [{"age" 20 "name" "jack"} {"age" 30 "name" "tom"}])
(assert (= ["tom"] (hsearch arr "#(age>20).name")))

;; escape
(def h2 {"fav.movie" "Deer Hunter"})
(assert (= "Deer Hunter" (hsearch h2 "fav\\.movie")))

(def h3 {"a1" [{"name" "jack" "age" "20" "height" 174.3} {"name" "tom" "age" "30" "height" 160.3}]})
(assert (= ["tom"] (hsearch h3 "a1.#(age>20).name")))
(assert (= ["jack" "tom"] (hsearch h3 "a1.#(age>=20).name")))
(assert (= ["jack"] (hsearch h3 "a1.#(age<30).name")))
(assert (= ["jack" "tom"] (hsearch h3 "a1.#(age<=30).name")))
