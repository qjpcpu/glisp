(def global.var1 1)
(assert (= 1 global.var1))

(defn myfn1 []
  (assert (= global.var1 1))
  (def global.var1 2)
  (assert (= global.var1 2)))

(myfn1)
(assert (= 1 global.var1))

(defn myfn2 []
  (assert (= global.var1 1))
  (set! global.var1 2)
  ;; local
  (set! global.var2 1)
  (assert (= global.var1 2))
  (assert (= global.var2 1)))

(myfn2)
(assert (= 2 global.var1))
