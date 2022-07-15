;; overwrite builtin functions
(defn + [a b] (* a b))
(assert (= 6 (+ 2 3)))
