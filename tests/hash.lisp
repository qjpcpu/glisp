(def h {'a 21 2 5 "c" #c})

; testing hget
(assert (= 21 (hget h 'a)))
(assert (= 5 (hget h 2)))
(assert (= #c (hget h "c")))
; default value
(assert (= 0 (hget h 22 0)))

; testing set
(hset! h #a 3)
(assert (= 3 (hget h #a)))

; testing confict resolution
; b and its symbol number have same hash value
(hset! h 'b 13)
(hset! h (symnum 'b) 42)

(assert (= 13 (hget h 'b)))
(assert (= 42 (hget h (symnum 'b))))

; testing delete
(hdel! h 'a)
; 'a should be gone
(assert (= 0 (hget h 'a 0)))
(hdel! h (symnum 'b))
; b should still be in there
(assert (= 13 (hget h 'b)))

(assert (hash? h))
(assert (empty? {}))
(assert (not (empty? h)))

(assert (exist? {1 "a"} 1))
(assert (not (exist? {1 "a"} 2)))
(assert (not (exist? {} 2)))
(assert (exist? {1 '()} 1))

(assert (= 1 (len {'a 1})))

(def h2 {true false false true})
(assert (exist? h2 true))
(assert (exist? h2 false))
(assert (= false (hget h2 true)))
(assert (= true (hget h2 false)))
(assert (= 1 (hget {18446744073709551615 1} 18446744073709551615)))

(def hh {'a 1 'b 2 'c 3 'd 4})
(def hhres (filter (fn [kv] (not= 0 (mod (cdr kv) 2))) hh))
(assert (= 2 (len hhres)))

(def h3 {'a 1 'b 2})
(assert (= 1 (len (hdel! h3 'a))))
(assert (= 2 (len (hset! h3 'a 1))))

(assert (= [1 2]
           (list-to-array (map #(cdr %1) {"a" 1 "b" 2}))))

(def h4 {"a" 1 'b 2})
(assert (= 1 (:a h4)))
(assert (= 2 (:b h4)))
(assert (= 3 (:c h4 3)))
