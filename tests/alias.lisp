(alias plus +)
(defn strange_name [a b] (+ (plus a  b) 1))
(alias plus2 strange_name)
(assert (= 3 (plus 1 2)))
(assert (= 4 (plus2 1 2)))

(alias UPPER str/upper)
(assert (= "ABC" (UPPER "abc")))
(assert (= "ABC" (str/upper "abc")))

(defn user-define-function [arg] (+ arg 1))

(assert (= 2 (user-define-function 1)))
(alias udf user-define-function)
(assert (= 2 (udf 1)))



;; overwrite function
(defn add1 [arg] (+ arg 1))
(defn add1 [arg] (+ arg 2))
(assert (= 3 (add1 1)))
