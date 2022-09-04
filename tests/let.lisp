(defn make-let-binding-args []
  ['a 100 'b 200])

(defn make-let-binding-args2 []
  {'a 100 'b 200})

(let (make-let-binding-args)
  (assert (= 100 a))
  (assert (= 200 b)))

(let (make-let-binding-args2)
  (assert (= 100 a))
  (assert (= 200 b)))
