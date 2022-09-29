(defn recursive-call [a & b]
  (cond (nil? b) a (recursive-call a)))

(assert (= 100 (recursive-call 100)))
(assert (= 100 (recursive-call 100 1 2 3)))

(defn normal-recursive-call [a & b]
  (cond (nil? b) a (+ 1 (normal-recursive-call a))))

(assert (= 100 (normal-recursive-call 100)))
(assert (= 101 (normal-recursive-call 100 1 2 3)))
