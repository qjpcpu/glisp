(assert (= 3
  ((fn [a b] a) 3 2)))

(defn add4 [a] (+ a 4))

(assert (= 7 (add4 3)))

; testing recursion
(defn fact [n]
  (cond (= n 0) 1 (* n (fact (- n 1)))))

(assert (= 120 (fact 5)))
(assert (= 3628800 (fact 10)))

(defn sum [l]
  (cond (empty? l)
    0 (+ (car l) (sum (cdr l)))))

(assert (= 0 (sum [])))
(assert (= 6 (sum [1 2 3])))

; testing tail recursion
(defn fact-tc [n accum]
  (cond (= n 0) accum
    (let [newn (- n 1)
          newaccum (* accum n)]
      (fact-tc newn newaccum))))

(assert (= 120 (fact-tc 5 1)))
(assert (= 3628800 (fact-tc 10 1)))

(defn sum-tc [l a]
  (cond (empty? l)
    a (sum-tc (cdr l) (+ a (car l)))))

(assert (= 0 (sum-tc [] 0)))
(assert (= 6 (sum-tc [1 2 3] 0)))

; testing anonymous dispatch
((fn [a] (assert (= a 0))) 0)

(assert (= "list" (type (read "()"))))
(assert (= 2 (apply '+ [1 1])))

(defn sub [f a b] (f a b))
(assert (= 1 (sub -  2 1)))

;; dynamic function name
(defn (symbol (concat "ab" "c")) [] "abc")
(assert (= "abc" (abc)))

(assert (= [1 2] (reject #(> % 2) [1 2 3 4])))
