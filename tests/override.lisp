(defn orig-fun [a b] (+ a b))

(override orig-fun (fn [a b] (+ 1 (orig-fun a b))))

(assert (= 6 (orig-fun 2 3)))
(assert (= 6 (orig-fun 2 3)))

;; override again
(override orig-fun (fn [a b] (+ 1 (orig-fun a b))))
(assert (= 7 (orig-fun 2 3)))
(assert (= 7 (orig-fun 2 3)))


(override - (fn [a b]
                        (cond (or (list? a) (array? a)) (list/complement a b) (- a b))))

(assert (= 1 (- 3 2)))
(assert (= [1] (- [1 2 3] [2 3])))
(assert (= '(1) (- '(1 2 3) '(2 3))))

(override len (fn [x] 100))
(assert (= 100 (len [1])))
