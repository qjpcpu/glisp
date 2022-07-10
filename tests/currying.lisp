(defn sum3 [ a b c ] (+ a b c))

(def g0 (currying sum3 3))
(assert (= "function" (typestr (g0))))

(def g1 (g0 1))
(assert (= "function" (typestr (g1))))
(assert (= 11 (g1 4 6)))

(def g2 (g1 2))
(assert (= "function" (typestr (g2))))

(def g3 (g2 3))
(assert (= "int" (typestr g3)))
(assert (= 6 g3))
(assert (= 13 (g2 10)))
(assert (= 13 (g2 10)))

;;; currying with lambda
(assert (= 6
           ((((currying (fn [a b c] (+ a b c)) 3) 1) 2) 3)
        ))


;;;overrying and overwrite
(defn add1 [a b] (+ a b))
(def add1 ((currying add1 2) 100))
(assert (= 101 (add1 1)))

;; currying with initial arguments
(def c1 (currying sum3 3 10 100))
(assert (= 111 (c1 1)))
(assert (= 112 (c1 2)))
