(assert (= '(2 3 4)
           (map (fn [number] (+ number 1)) '(1 2 3))
           ))

(assert (nil? (map (fn [e] 100) '())))
(assert (nil? (flatmap (fn [e] 100) '())))
