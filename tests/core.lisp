(assert (nil? (sort #(< %1 %2) nil)))
(assert (nil? (sort nil)))
(assert (= [1 2 3] (sort [3 2 1])))
(assert (= '(1 2 3) (sort '(3 2 1))))

(assert (= [2 3 1] (sort #(= % 2) [3 2 1])))
