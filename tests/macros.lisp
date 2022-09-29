(def l '(1 2 3))
(def b 4)

(assert (= `(0 ~@l ~b) '(0 1 2 3 4)))

(defmac when [predicate & body]
  `(cond ~predicate
      (begin
        ~@body) '()))

(assert (nil? (when false 'c)))
(assert (= 'a (when true 'c 'b 'a)))

(assert (=
         '(cond false (begin (quote c)) (quote ()))
         (macexpand (when false 'c))))
(assert (=
         '(cond true (begin (quote c) (quote b) (quote a)) (quote ()))
         (macexpand (when true 'c 'b 'a))))

(defmac if [& body]
  `(cond ~@body))

(assert (= 'a (if true 'a 'b)))

(assert (= 'b (if (> 1 2) 'a 'b)))

(defmac unless [predicate & body]
 `(cond (not ~predicate)
     (begin
       ~@body) '()))

(assert (= 'a (unless false 'a)))
(assert (= 'a (unless (< 2 1) 'a)))

(defn test-begin-in-condition []
  (cond true
        (begin
         (+ 1 2)
         (* 2 3))
        '()))

(assert (= 6 (test-begin-in-condition)))
