(defn add3 [a] (+ a 3))

(assert (= '(4 5 6) (map add3 '(1 2 3))))
(assert (= [4 5 6] (map add3 [1 2 3])))

(assert (= (apply (fn [a b] (* 2 a b)) [1 2]) 4))
(assert (= (apply (fn [a b] (* 2 a b)) '(1 2)) 4))

;; map nil
(assert (= ['() '() '()] (map (fn [a] '()) [1 2 3])))
(assert (= (list '() '() '()) (map (fn [a] '()) '(1 2 3))))

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
            ;; append char ! to string
            (fn [a] (append a #!))
            ;; convert integer to string
            string
            ;; add 1 to every integer
            (fn [a] (+ a 1))
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

(defn fn1 [] 1)
(defn fn2 [] 2)
(defn fn3 [] 3)

(assert (= [1 2 3] (map (fn [f] (f)) [fn1 fn2 fn3])))

;; no input output compose
(assert (nil? ((compose (fn [e] '()) (fn [] '()) ))))

(def compose_lambda (compose
 #(filter #(not= 2 %1) %1)
 #(map int %1)
 #(flatmap #(str/split %1 " ") %1)))

(assert (= [1 3 4 5 6] (compose_lambda ["1 2 3" "4 5 6"])))

(assert (= [1 3 4 5 6]
            (->> ["1 2 3" "4 5 6"]
                 (flatmap #(str/split %1 " "))
                 (map int)
                 (filter #(not= 2 %1)))))

(def compose_lambda (compose
 #(filter #(not= 2 %) %)
 #(map int %)
 #(flatmap #(str/split % " ") %)))

(assert (= [1 3 4 5 6] (compose_lambda ["1 2 3" "4 5 6"])))

(assert (= [1 3 4 5 6]
           (->> ["1 2 3" "4 5 6"]
                (flatmap #(str/split % " "))
                (map int)
                (filter #(not= 2 %)))))

;; thread last
(assert (= "321INIT"
        (->> "INIT" (concat "1") (concat "2") (concat "3"))))

(assert (= "321INIT"
        (->> ((fn [] "INIT")) (concat "1") (concat "2") (concat "3"))))

(assert (= 20
           (->> [1 2 3 4 5]
                (#(map #(+ 1 %) %)) ; double parentheses
                (apply +))))

(assert (= 20
           (->> [1 2 3 4 5]
                ((fn [e] (map #(+ 1 %) e)))
                (apply +))))


;; thread first
(assert (= "INIT123"
        (-> "INIT" (concat "1") (concat "2") (concat "3"))))

(assert (= 1 (->> 1)))
(assert (= 1 (-> 1)))

;; https://clojure.org/guides/learn/functions#_gotcha
(assert (= [100] (#([%]) 100)))
(assert (= 100 (#(%) 100)))
(assert (= true (#(%) true)))
(assert (= 100 (hget (#({"a" %1}) 100) "a")))
(assert (= nil (#(%))))
