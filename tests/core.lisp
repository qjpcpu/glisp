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
