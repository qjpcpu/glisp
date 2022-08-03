(let [a (gensym) b (gensym)]
     (assert (not= a b)))

(defmac xyz []
  (gensym)
  '())

(xyz)

(let [a (gensym) b (gensym)]
     (assert (not= a b)))
