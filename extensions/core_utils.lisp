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

;; partial
(defn partial [function & args]
  (fn [& extra]
      (let [all (concat args extra)]
           (apply function all))))

;; override
(defmac override [function new_function]
  `(def ~function (let [~function ~function] ~new_function)))
