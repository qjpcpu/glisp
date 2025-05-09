;; not-nil?
(defn not-nil? [x]
  "Usage: (not-nil? x)
Return true if x is not nil."
  (not (nil? x)))

(defn not-empty? [x]
  "Usage: (not-empty? x)
Return true if x is not empty."
  (not (empty? x)))

(defn core/id [x]
  "Usage: (core/id x)
Always return x"
  x)

(defn core/const [x]
  "Usage: (core/const x)
Return a function which always return x."
  (fn [& anything] x))

(defn core/dummy [x]
  "Usage: (core/dummy x)
Always return nil."
  nil)

;; (car (cdr x))
(defn cadr [x]
  "Usage: (cadr x)
Equivalent to (car (cdr x))"
  (car (cdr x)))

;; alias function
(defmac alias [new old]
  "Usage: (alias new_name orig_name)

Set alias for function."
  `(defn ~new [& body] (apply ~old body)))

;; currying
(defn core/__gen_curry [function arg_count iargs args]
  (cond (>= (+ (len iargs) (len args)) arg_count)
        (apply function (concat iargs args))
        (fn [& body]
          (core/__gen_curry function arg_count (concat iargs args) body))))

(defn currying [function arg_count & args]
  "Usage: (currying f total-arguments-count & args)

Transform a function that takes multiple arguments into a function for which some of the arguments are preset."
  (core/__gen_curry function arg_count args '()))

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

;; thread first
(defmac -> [init-value & functions]
  "Usage: (-> x & forms)

Threads the expr through the forms. Inserts x as the
second item in the first form. If there are more forms, inserts the first form as the
second item in second form, etc.
"
  (foldl (fn [expr acc] (concat (list (car expr)) (list acc) (cdr expr))) init-value functions))

(defmac some-> [init-value & functions]
  "Usage: (some-> expr & forms)
When expr is not nil, threads it into the first form (via ->),
and when that result is not nil, through the next etc"
  (foldl (fn [expr acc]
           (let* [x (gensym) form (concat (list (car expr)) (list x) (cdr expr))]
                 `(let [~x ~acc] (cond (nil? ~x) nil ~form)))) init-value functions))

(defmac ->> [init-value & functions]
  "Usage: (->> x & forms)

Threads the expr through the forms. Inserts x as the
last item in the first form. If there are more forms, inserts the first form as the
last item in second form, etc.
"
  (foldl (fn [expr acc] (concat expr (list acc))) init-value functions))

(defmac some->> [init-value & functions]
  "Usage: (some->> expr & forms)
When expr is not nil, threads it into the first form (via ->>),
and when that result is not nil, through the next etc"
  (foldl (fn [expr acc]
           (let* [x (gensym) form (concat expr (list x))]
                 `(let [~x ~acc] (cond (nil? ~x) nil ~form))))
         init-value functions))

(defmac inverse-> [arg & args]
  "Usage: (inverse-> f & args)
Transform a thread first form to a thread last form.
e.g. transform (inverse-> arg1 fn arg2 arg3) to (fn arg2 arg3 arg1)"
  (let [a (concat args (list arg))]
    `(~@a)))

(defmac inverse->> [f & args]
  "Usage: (inverse->> f & args)
Transform a thread last form to a thread first form.
e.g. transform (inverse->> fn arg1 arg2 arg3) to (fn arg3 arg1 arg2)"
  (let* [arr (list-to-array args) n (len arr) arg (aget arr (- n 1)) args2 (cons arg (array-to-list (slice arr 0 (- n 1))))]
        `(~f ~@args2)))

(defn array-to-list [arr]
  "Usage: (array-to-list arr)"
  (cond (empty? arr) nil
        (list? arr) arr
        (cons (aget arr 0) (array-to-list (slice arr 1)))))

(defn list-to-array [x]
  "Usage: (list-to-array x)"
  (foldl #(append %2 %1) [] x))

(defmac case [e & clauses]
  "Usage: (case e & clauses)
Takes an expression, and a set of clauses.
 Each clause can take the form of either:
 test-constant result-expr
 (test-constant1 ... test-constantN)  result-expr
If the expression is equal to a
test-constant, the corresponding result-expr is returned. A single
default expression can follow the clauses, and its value will be
returned if no clause matches.

example:
;; return result x when value x matched
;; return nil when nothing matched
(case expr
      value1 result1
      value2 result2)

;; return result x when value x matched
;; return default-result when nothing matched
(case expr
      value1 result1
      value2 result2
      default-result)
"
  (let* [x (gensym) expr (->> (cond (= 0 (mod (len clauses) 2)) (concat clauses (list nil)) clauses)
                              (stream)
                              (partition 2)
                              (flatmap (fn [pair] (cond (= 2 (len pair)) (list (list '= x (car pair)) (car (cdr pair))) pair)))
                              (realize))]
        `(let [~x ~e] (cond ~@expr))))

(defn foreach [f coll]
  "Usage: (foreach f coll)

Apply f to each item of coll"
  (let [f2 (fn [e] (f e) e)] (map f2 coll)))

(defmac when [predicate & body]
  "Usage: (when test & body)
Evaluates test. If logical true, evaluates body in an implicit begin."
  `(cond ~predicate
         (begin
          ~@body) '()))

;; well, maybe somebody likes if more than when
(defmac if [predicate & body]
  "Usage: (if test & body)
Evaluates test. If logical true, evaluates body in an implicit begin."
  `(cond ~predicate
         (begin
          ~@body) '()))

(defmac unless [predicate & body]
  "Usage: (unless test & body)
Evaluates test. If logical false, evaluates body in an implicit begin."
  `(cond (not ~predicate)
         (begin
          ~@body) '()))

(defn index-of [f x]
  "Usage: (index-of f x)
Find index of x by f, return -1 if not found.
f can be a predicate function or a constant literal,
x can be stream or list or array.
e.g.
(index-of \"lisp\" [\"go\" \"lisp\"]) ; return 1
(index-of #(= % \"lisp\") [\"go\" \"lisp\"]) ; return 1
"
  (let [fx (cond (function? f) f (fn [v] (= v f)))]
    (->> (zip (range) (stream x))
         (filter (fn [e] (fx (car (cdr e)))))
         (map #(car %))
         (realize)
         (car)
         ((fn [e] (cond (nil? e) -1 e))))))

(defn nth [n coll]
  "Usage: (nth n coll)

Return n-th element of array/list/stream, n start from 0"
  (some->> (cond (< n 0) nil coll)
           (stream)
           (drop n)
           (take 1)
           (realize)
           (car)))

(defn reverse [x]
  "Usage: (reverse x)

Reverse list/array/string/stream."
  (let [res (->> (stream x)
                 (foldl cons '()))]
    (cond (array? x) (list-to-array res)
          (string? x) (string res)
          (bytes? x) (bytes (string res))
          (stream? x) (stream res)
          res)))

(defn list/complement [a b]
  "Usage: (list/complement a b)

Return coll which elements belongs to a but not belongs to b.
Result coll = a \\ b."
  (let [h (foldl (fn [e acc] (hset! acc e '()) acc) {} b)]
    (filter (fn [e] (not (exist? h e))) a)))

(override - (fn [& args]
              (let [length (len args)]
                (cond (and (>= length 2) (list? (car args)) (list? (car (cdr args)))) (foldl (fn [b a] (list/complement a b)) (car args) (cdr args))
                      (and (>= length 2) (array? (car args)) (array? (car (cdr args)))) (foldl (fn [b a] (list/complement a b)) (car args) (cdr args))
                      (and (>= length 2) (stream? (car args)) (stream? (car (cdr args))))
                      (foldl (fn [b a]
                               (let [h (core/__stream2hash b)] (reject #(exist? h %) a))) (car args) (cdr args))
                      (apply - args)))))

(defn list/intersect [a b]
  "Usage: (list/intersect a b)
Return coll which elements belongs to a and b."
  (let [h (foldl (fn [e acc] (hset! acc e '()) acc) {} b)]
    (filter (fn [e] (exist? h e)) a)))

(defn core/__stream2hash [s]
  (foldl (fn [e acc] (hset! acc e true)) {} s))

(defn uniq [& a]
  "Usage: (uniq a)
(uniq fn a)
Drop duplicate elements of list/array/stream a."
  (cond (= 1 (len a)) (core/__uniq-by core/id (car a))
        (and (= 2 (len a)) (function? (car a))) (core/__uniq-by (car a) (cadr a))
        (assert false (sprintf "bad pattern (uniq%v)" (foldl (fn [e acc] (concat acc " " (type e))) "" a)))))

(defn core/__uniq-by [f a]
  "Usage: (core/__uniq-by fn a)
Drop duplicate elements of list/array/stream a."
  (let* [h {} ret (foldl (fn [e acc] (let [ex (f e)] (cond (exist? h ex) acc (begin (hset! h ex 1) (concat acc (list e)))))) '() a)]
        (cond (list? a) ret
              (array? a) (list-to-array ret)
              (stream ret))))

(defn reject [pred coll]
  "Usage: (reject pred coll)
Returns a sequence of the items in coll for which
(pred item) returns logical false. pred must be free of side-effects.

When coll is hash, pred function should take 2 arguments, which are hash key-value pair.
"
  (filter (fn [e] (not (pred e))) coll))

(defn flatten [x]
  "Usage: (flatten coll)
Takes any nested combination of sequential things (lists, vectors,
etc.) and returns their contents as a single, flat foldable
collection."
  (flatmap (fn [e] e) x))

(override take (fn [f st]
                 (case (type st)
                   "list" (->> (stream st) (take f) (realize))
                   "array" (->> (stream st) (take f) (realize) (list-to-array))
                   (take f st))))

(override drop (fn [f st]
                 (case (type st)
                   "list" (->> (stream st) (drop f) (realize))
                   "array" (->> (stream st) (drop f) (realize) (list-to-array))
                   (drop f st))))

(defn max [& numbers]
  "Usage: (max a b c)
Return maximun value."
  (cond (= 0 (len numbers)) 0
        (foldl (fn [e acc] (cond (> e acc) e acc)) (car numbers) numbers)))

(defn min [& numbers]
  "Usage: (min a b c)
Return minimun value."
  (cond (= 0 (len numbers)) 0
        (foldl (fn [e acc] (cond (< e acc) e acc)) (car numbers) numbers)))
