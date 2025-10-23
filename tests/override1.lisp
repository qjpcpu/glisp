(defn minus [a b] (- a b))
                                        ; test minus
(assert (= 1 (minus 3 2)))
(assert (= 1 (- 3 2)))
                                        ; override -
(override minus (fn [a b] (- a b 1)))
                                        ; test new -
(assert (= 0 (minus 3 2)))
