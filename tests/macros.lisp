(def l '(1 2 3))
(def b 4)

(assert (= `(0 ~@l ~b) '(0 1 2 3 4)))

(assert (nil? (when false 'c)))
(assert (= 'a (when true 'a)))

(assert (=
         '(cond false (begin (quote c)) (quote ()))
         (macexpand (when false 'c))))
(assert (=
         '(cond true (begin (quote c) (quote b) (quote a)) (quote ()))
         (macexpand (when true 'c 'b 'a))))


(assert (= 'b (if true 'a 'b)) )

(assert (= nil (if (> 1 2) 'a 'b)))

(assert (= 'a (unless false 'a)))
(assert (= 'a (unless (< 2 1) 'a)))

(defn test-begin-in-condition []
  (cond true
        (begin
         (+ 1 2)
         (* 2 3))
        '()))

(assert (= 6 (test-begin-in-condition)))

(assert (= 1 (if true 1)))
(assert (= nil (if false 1)))

(assert (= 1 (when true 1)))
(assert (= nil (when false 1)))

(assert (= 1 (unless false 1)))
(assert (= nil (unless true 1)))

(assert (nil? (some-> nil)))
(assert (nil? (some-> 1 (+ 2) ((fn [e] nil)) (/ 1 0))))
(assert (nil? (some->> 1 (+ 2) ((fn [e] nil)) (/ 1 0))))
