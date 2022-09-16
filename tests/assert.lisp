(def h {})
(assert true (begin (hset! h "a" 1) "should not execute"))
