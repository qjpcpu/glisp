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

Return n-th element of array/list/stream."
  (some->> (cond (< n 0) nil coll)
           (stream)
           (drop n)
           (take 1)
           (realize)
           (car)))

(defn reverse [x]
  "Usage: (reverse x)

Reverse list/array/stream."
  (let [res (->> (zip (range) (stream x))
                 (realize)
                 (sort #(> (car %1) (car %2)))
                 (map #(car (cdr %))))]
    (cond (array? x) (list-to-array res)
          (string? x) (string res) res)))
