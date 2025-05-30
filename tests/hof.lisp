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


(assert (= 4 (foldl max 0 '(1 2 3 4) )))
(assert (= 4 (foldl max 0 '[1 2 3 4] )))

(assert (= 1 (foldl min 1000 '(1 2 3 4) )))
(assert (= 1 (foldl min 1000 '[1 2 3 4] )))

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

(assert (= 606 (foldl (fn [kv acc] (+ (car kv) (cdr kv) acc)) 0 {1 100 2 200 3 300})))
(assert (= 0 (foldl (fn [kv acc] (+ (car kv) (cdr kv) acc)) 0 {})))
(assert (= [1 100 2 200 3 300] (foldl (fn [kv acc] (append acc (car kv) (cdr kv))) [] {1 100 2 200 3 300})))

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

;; some version
(assert (= "321INIT"
        (some->> "INIT" (concat "1") (concat "2") (concat "3"))))

(assert (= "321INIT"
        (some->> ((fn [] "INIT")) (concat "1") (concat "2") (concat "3"))))

(assert (= 20
           (->> [1 2 3 4 5]
                (#(map #(+ 1 %) %)) ; double parentheses
                (apply +))))

(assert (= 20
           (->> [1 2 3 4 5]
                ((fn [e] (map #(+ 1 %) e)))
                (apply +))))

;; some version
(assert (= 20
           (some->> [1 2 3 4 5]
                (#(map #(+ 1 %) %)) ; double parentheses
                (apply +))))

(assert (= 20
           (some->> [1 2 3 4 5]
                ((fn [e] (map #(+ 1 %) e)))
                (apply +))))


;; thread first
(assert (= "INIT123"
        (-> "INIT" (concat "1") (concat "2") (concat "3"))))

(assert (= 1 (->> 1)))
(assert (= 1 (-> 1)))
(assert (= 1 (some->> 1)))
(assert (= 1 (some-> 1)))

;; https://clojure.org/guides/learn/functions#_gotcha
(assert (= [100] (#([%]) 100)))
(assert (= 100 (#(%) 100)))
(assert (= true (#(%) true)))
(assert (= 100 (hget (#({"a" %1}) 100) "a")))
(assert (= nil (#(%))))

(assert (= [100 200 300] (flatten  [[100] [200 300]])))
(let [l (flatten (list (list 1) (list 2)))]
                 (assert (= 2 (len l)))
                 (assert (= 1 (car l)))
                 (assert (= 2 (car (cdr l))))
                 )

;; filter hash
(assert (= 2 (->> {"a" 1 "b" 2}
                  (filter (fn [kv] (= 2 (cdr kv))))
                  (foldl #(cdr %1) 0))))

(assert (= 2 (some->> {"a" 1 "b" 2}
                  (filter (fn [kv] (= 2 (cdr kv))))
                  (foldl #(cdr %1) 0))))

(assert (= [1 2] (list-to-array (->> {"a" 1 "b" 2}
                  (flatmap #([(cdr %1)]))))))

(assert (= [1 2] (list-to-array (->> {"a" 1 "b" 2}
                  (flatmap #(list (cdr %1)))))))


;; composite thread first/last macro
(assert (= "CBAabcd"
           (-> "a"
               (concat "b")
               (concat "c")
               (->>
                (concat "A")
                (concat "B")
                (concat "C"))
               (concat "d"))))

(assert (= "CBAabcd"
           (some-> "a"
               (concat "b")
               (concat "c")
               (->>
                (concat "A")
                (concat "B")
                (concat "C"))
               (concat "d"))))

(assert (= nil (case "no cases")))
(assert (= "match1" (case 1 0 "match0" 1 (concat "match" "1"))))
(assert (= "default" (case 33 0 "match0" 1 (concat "match" "1") "default")))
(assert (= nil (case "orig" "a" "match0" "b" (concat "match" "1"))))
(assert (= [1 2 3] (foreach string [1 2 3])))
