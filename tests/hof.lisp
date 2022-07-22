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
