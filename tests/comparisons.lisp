(assert (< 0 1))
(assert (< 3.2 4.7))
(assert (> 1.6 -3.1))
(assert (= 1.1 1.1))
(assert (<= 1 1))
(assert (>= 1 1))

(assert (< "a" "b"))
(assert (> #b #a))
(assert (< "abc" "abcd"))

(assert (< '(1 2 3) '(1 2 4)))

(assert (!= 1 2))

(assert (= 97 #a))

(assert (< '() 1))
(assert (> 1 '()))
(assert (= '() '()))
(assert (= #a 97))
(assert (= #a 97.0))
(assert (= 0B6869 "hi"))
(assert (> true false))
(assert (< false true))
