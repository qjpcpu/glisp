;; alias function
(defmac alias [new old]
 `(defn ~new [& body] (apply ~old body))
)

;; currying
(defn __gen_curry [function arg_count iargs args]
  (cond (>= (+ (len iargs) (len args)) arg_count)
            (apply function (concat iargs args))
        (fn [& body]
          (__gen_curry function arg_count (concat iargs args) body))))

(defn currying [function arg_count & args]
  (__gen_curry function arg_count args '())
)