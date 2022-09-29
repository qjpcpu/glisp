;; define global variables
(def nil '())

;; alias function
(defmac alias [new old]
  "Usage: (alias new_name orig_name)

Set alias for function."
 `(defn ~new [& body] (apply ~old body))
)

;; currying
(defn __gen_curry [function arg_count iargs args]
  (cond (>= (+ (len iargs) (len args)) arg_count)
            (apply function (concat iargs args))
        (fn [& body]
          (__gen_curry function arg_count (concat iargs args) body))))

(defn currying [function arg_count & args]
  "Usage: (currying f total-arguments-count & args)

Transform a function that takes multiple arguments into a function for which some of the arguments are preset."
  (__gen_curry function arg_count args '())
)

;; partial
(defn partial [function & args]
  "Usage: (partial f & more)

Takes a function f and fewer than the normal arguments to f. Returns a function that takes a variable number of additional arguments. When called, the returned function calls f with the original arguments plus the additional arguments.

((partial f a b) c d) => (f a b c d)
"
  (fn [& extra]
      (let [all (concat args extra)]
           (apply function all))))

;; override
(defmac override [function new_function]
  "Usage: (override original_func_symbol anonymous_function_implement)

example:
;; original function addone
(defn addone [a] (+ a 1))

;; override function addone, printing input before calculate
(override addone (fn [a]
    (println a)
    (addone a) ;; invoke original function
))
"
  `(def ~function (let [~function ~function] ~new_function)))

(defmac doc [name]
  "Usage: (doc f)

Display document of function."
  `(println (__doc__ (quote ~name))))
