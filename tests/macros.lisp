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

;; testing macro expanding
(assert (= "(defn my-plus [& body] (apply + body))" (sexp-str (macexpand (alias my-plus +)))))
(assert (= "(def + (let [+ +] (fn [a b] (+ a b))))" (sexp-str (macexpand (override + (fn [a b] (+ a b)))))))
(assert (= "(* (+ 0 1) 2)" (sexp-str (macexpand (-> 0 (+ 1) (* 2))))))
(assert (= "(* 2 (+ 1 0))" (sexp-str (macexpand (->> 0 (+ 1) (* 2))))))

;; fuzzy macro
(defmac #`always-return-\d+` [& args] 1024)
(assert (= 1024 (always-return-1024)))
(assert (= 1024 (always-return-1 "useless")))


(defmac #`^Xreturn-number-\d+$` [name]
        (let [arr (str/split (string name) "-")]
                             (int (aget arr 2))))
(assert (= 1024 (Xreturn-number-1024)))
(assert (= 100 (Xreturn-number-100)))

(defmac #`^X[a-zA-Z]+$` [name h]
        (let [f (-> name (str/trim-prefix "X"))]
             `(hget ~h ~f nil)))

(def ash {"a" 1 "b" 2 "c" 3 "d" 4})
(assert (= 1 (Xa ash)))
(assert (= 2 (Xb ash)))
(assert (nil? (Xx ash)))
(assert (= 1 (-> ash (Xa))))

(defmac #`^Hello\d+$` [n] `(+ 1 1))
(assert (= 2 (Hello1)))
;; overwrite fuzzy macro
(defmac #`^Hello\d+$` [n] `(+ 1 2))
(assert (= 3 (Hello1)))
;; fuzzy not work, match previous
(defmac #`^Hello1$` [n] `(+ 1 3))
(assert (= 3 (Hello1)))
