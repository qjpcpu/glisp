(assert (= "function" (typestr #'+)))

(assert (= 3 (apply + '(1 2))))
(assert (= 3 (apply #'+ '(1 2))))


(assert (= 3 (apply (fn [a b] (+ a b)) '(1 2))))
(assert (= 3 (apply #'(fn [a b] (+ a b)) '(1 2))))
(assert (= 3 (#'(fn [a b] (+ a b)) 1 2)))




(assert (= "#'" (str #\')))
(assert (= "#(" (str #\()))
(assert (= "#)" (str #\))))
(assert (= "#[" (str #\[)))
(assert (= "#]" (str #\])))
(assert (= "##" (str #\#)))
(assert (= "#~" (str #\~)))
(assert (= "#`" (str #\`)))
(assert (= "#," (str #,)))
(assert (= "#@" (str #@)))
(assert (= "#," (str #,)))