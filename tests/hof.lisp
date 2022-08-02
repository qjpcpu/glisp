(defn add3 [a] (+ a 3))

(assert (= '(4 5 6) (map add3 '(1 2 3))))
(assert (= [4 5 6] (map add3 [1 2 3])))

(assert (= (apply (fn [a b] (* 2 a b)) [1 2]) 4))
(assert (= (apply (fn [a b] (* 2 a b)) '(1 2)) 4))


(assert (= 10 (foldl + 0 [1 2 3 4])))
(assert (= 24 (foldl * 1 [1 2 3 4])))

(assert (= 10 (foldl + 0 '(1 2 3 4) )))
(assert (= 24 (foldl * 1 '(1 2 3 4) )))


(assert (= 1 (foldl + 1 '() )))
(assert (= 1 (foldl + 1 [])))

(assert (= [1 3] (filter (fn [a] (= 1 (mod a 2))) [1 2 3 4])))
(assert (= '(1 3) (filter (fn [a] (= 1 (mod a 2))) '(1 2 3 4))))

(assert (= [] (filter (fn [a] (= 1 (mod a 2))) [])))
(assert (= '() (filter (fn [a] (= 1 (mod a 2))) '())))

(assert (= [] (filter (fn [a] false) [1 2 3 4])))
(assert (= '() (filter (fn [a] false) '(1 2 3 4))))

(defn max [a b]
  (cond (> a b) a b))


(assert (= 4 (foldl max 0 '(1 2 3 4) )))
(assert (= 4 (foldl max 0 '[1 2 3 4] )))

(assert (= ["101!" "201!" "301!" "401!"] (map
          (compose
            ;; add 1 to every integer
            (fn [a] (+ a 1))
            ;; convert integer to string
            str
            ;; append char ! to string
            (fn [a] (append a #!))
            )
          [100 200 300 400]))
        )
(assert (= [1 101 2 102 3 103] (flatmap (fn [a] [a (+ a 100)]) [1 2 3])))
(assert (= [100 101 300 301] (flatmap (fn [a] (cond (= a 200) [] [a (+ 1 a)])) [100 200 300])))
(assert (= [200 201 300 301] (flatmap (fn [a] (cond (= a 100) [] [a (+ 1 a)])) [100 200 300])))
(assert (= [200 201 300 301] (flatmap (fn [a] (cond (= a 100) '() [a (+ 1 a)])) [100 200 300])))
(assert (= [] (flatmap (fn [a] '()) [100 200 300])))

(assert (= '(100 101) (flatmap (fn [a] (list a (+ 1 a))) '(100))))
(assert (= '(100 101 200 201 300 301) (flatmap (fn [a] (list a (+ 1 a))) '(100 200 300))))
(assert (= '(100 101 300 301) (flatmap (fn [a] (cond (= a 200) '() (list a (+ 1 a)))) '(100 200 300))))
(assert (= '(200 201 300 301) (flatmap (fn [a] (cond (= a 100) '() (list a (+ 1 a)))) '(100 200 300))))
(assert (= '() (flatmap (fn [a] '()) '(100 200 300))))

(assert (= 606 (foldl (fn [k v acc] (+ k v acc)) 0 {1 100 2 200 3 300})))
(assert (= 0 (foldl (fn [k v acc] (+ k v acc)) 0 {})))
(assert (= [1 100 2 200 3 300] (foldl (fn [k v acc] (append acc k v)) [] {1 100 2 200 3 300})))
