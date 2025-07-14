(assert (nil? (sort #(< %1 %2) nil)))
(assert (nil? (sort nil)))
(assert (= [1 2 3] (sort [3 2 1])))
(assert (= '(1 2 3) (sort '(3 2 1))))

(assert (= [2 3 1] (sort #(= % 2) [3 2 1])))

(assert (= 3 (#'(-> "+" (symbol)) 1 2)))
(assert (= 3 (#'+ 1 2)))
(assert (nil? #'(symbol "xyzjjlk")))
(assert (= 3 (#'(#(symbol %) "+") 1 2)))

(assert (not-nil? +))

(assert (= "{\"a\":1,\"b\":2}" (json/stringify (union {"a" 1} {"b" 2}))))

(assert (nil? (:not-exist-property nil)))
(assert (nil? (:not-exist-property (car '()))))

(def test-json {"mylist" ["A" "B" "C"]})

(assert (= "B"
           (->> test-json
                (inverse->> json/query "mylist")
                (cadr))))

(assert (= "B"
           (-> test-json
               (json/query "mylist")
               (inverse-> filter #(= "B" %))
               (car))))

(assert (= '(1) (array-to-list '(1))))
(assert (= [1] (list-to-array [1])))

(assert (defined? len))
(assert (defined? (symbol (concat "l" "e" "n"))))
(assert (not (defined? (symbol (concat "l" "e" "ndxxx")))))
(assert (not (defined? "length_xxx")))

(assert (= (cons "a" 1) (car {"a" 1})))
(assert (nil? (car {})))

(assert (= "error" (type (error "test-error"))))
(assert (error? (error "test-error")))
