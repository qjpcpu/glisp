(assert (not (stream? nil)))
(assert (not (stream? [])))
(assert (not (stream? {})))
(assert (not (stream? "")))
(assert (not (stream? (bytes ""))))
(assert (not (stream? (list 1))))

(assert (stream? (stream nil)))
(assert (stream? (stream [])))
(assert (stream? (stream {})))
(assert (stream? (stream "")))
(assert (stream? (stream (bytes ""))))
(assert (stream? (stream (list 1))))

(assert (not (nil? (stream (range)))))
(assert (streamable? nil))
(assert (streamable? (range)))
(assert (streamable? []))
(assert (streamable? {}))
(assert (streamable? ""))
(assert (streamable? (bytes "")))
(assert (streamable? (list 1)))

(assert (not (stream? (my-counter 1))))
(assert (stream? (stream (my-counter 1))))

(let [counter (my-counter 100)]
     (def score (->> (stream counter)
          (filter #(= 0 (mod % 2)))
          (flatmap (fn [e] [-1024 e]))
          (map #(+ 100 %))
          (filter #(>= % 0))
          (take 3)
          (realize)))
     ;; return (102 104 106)
     (assert (= 3 (len score)))
     (assert (= 102 (car score)))
     (assert (= 104 (car (cdr score))))
     (assert (= 106 (car (cdr (cdr score)))))
     (assert (= 6 (get-my-counter counter)))
     )

(assert (= [10 11 12]
           (list-to-array
               (->> (stream (my-counter 100))
                    (map (fn [e] [1 e]))
                    (flatten)
                    (filter #(>= % 10))
                    (take 3)
                    (realize)))))

(assert (= 6
           (->> (stream (my-counter 100))
                (filter #(<= % 3))
                (foldl #(+ %1 %2) 0))))

(assert (= [10 11 12]
           (list-to-array
               (->> (stream (list 1 2 3 4 5 6 7 8 9 10 11 12 13))
                    (map (fn [e] [1 e]))
                    (flatten)
                    (filter #(>= % 10))
                    (take 3)
                    (realize)))))

(assert (= [10 11 12]
           (list-to-array
               (->> (stream [1 2 3 4 5 6 7 8 9 10 11 12 13])
                    (map (fn [e] [1 e]))
                    (flatten)
                    (filter #(>= % 10))
                    (take 3)
                    (realize)))))

(assert (= [#d #e #f]
           (list-to-array
               (->> (stream "abcdefghijklmn")
                    (filter #(exist? [#d #e #f] %))
                    (realize)))))

(assert (= [#d #e #f]
           (list-to-array
               (->> (stream (bytes "abcdefghijklmn"))
                    (filter #(exist? [#d #e #f] %))
                    (realize)))))

;; flatmap inner stream
(assert (= [1 2 3 1 2]
           (list-to-array
               (->> (stream [(stream (my-counter 3)) (stream (my-counter 2))])
                    (flatmap (fn [e] e))
                    (realize)))))

(assert (= [1 2 3 1 2]
           (list-to-array
               (->> (stream [(stream (my-counter 3)) (stream (my-counter 2))])
                    (flatten)
                    (realize)))))

(let [score (->> (stream (list 1 nil 2))
                 (flatmap (fn [e] [100 e]))
                 (realize))]
                 (assert (= [100 1 100 () 100 2] (list-to-array score))))

(assert (= [1 2]
           (list-to-array (realize (stream (cons 1 2))))))

(assert (= [1]
           (list-to-array (realize (stream (cons 1 nil))))))

(assert (nil? (realize (stream nil))))
(assert (nil? (realize (stream []))))
(assert (= "(stream [])" (sexp-str (stream []))))
(assert (= "(stream \"\")" (sexp-str (stream ""))))
(assert (= "(stream 0B)" (sexp-str (stream (bytes "")))))
(assert (= "(stream (1))" (sexp-str (stream (list 1)))))

(assert (= 3
           (->> (stream {"a" 1 "b" 2})
                (foldl #(+ (cdr %1) %2) 0))))

(assert (= 2
           (->> (stream {"a" 1 "b" 2})
                (filter #(= 2 (cdr %1)))
                (foldl #(+ (cdr %1) %2) 0))))

(assert (= 3
           (->> (stream {"a" 1 "b" 2})
                (map #(cdr %1))
                (foldl #(+ %1 %2) 0))))

(assert (= [1 2]
           (list-to-array
               (->> (stream [(stream (my-counter 100))])
                    (flatten)
                    (take #(< % 3))
                    (realize)))))

(assert (= [99 100]
           (list-to-array
               (->> (stream [(stream (my-counter 100))])
                    (flatten)
                    (drop #(<= % 98))
                    (realize)))))


(assert (= [99 100]
           (list-to-array
               (->> (stream [(stream (my-counter 100))])
                    (flatten)
                    (drop 98)
                    (realize)))))

(assert (= [1 2]
           (list-to-array
               (->> (stream [(stream (my-counter 100))])
                    (flatten)
                    (reject #(> % 2))
                    (realize)))))

;; range
(assert (= "(range)" (sexp-str (range))))
(assert (= "(range 0 100 1)" (sexp-str (range 100))))
(assert (= "(range 1 100 1)" (sexp-str (range 1 100))))
(assert (= "(range 1 100 2)" (sexp-str (range 1 100 2))))

(assert (= [0 1 2]
           (list-to-array
            (->> (range)
                 (take 3)
                 (realize)))))

(assert (= [0 1]
           (list-to-array
            (->> (range 2)
                 (take 3)
                 (realize)))))

(assert (= [1 2]
           (list-to-array
            (->> (range 1 3)
                 (realize)))))

(assert (= [1 4 7]
           (list-to-array
            (->> (range 1 10 3)
                 (realize)))))

(assert (= [3 3 3]
           (list-to-array
            (->> (range 3 10 0)
                 (take 3)
                 (realize)))))

(assert (= [42 43 44 45 46]
           (list-to-array
            (->> (range)
                 (drop 42)
                 (take 5)
                 (realize)))))
(assert (= nil (realize (range 0 0))))
(assert (= [0] (list-to-array (realize (range 0 1)))))


(defn concat-num [l]
  (str/join (list-to-array (map string l)) "-"))

;; partition
(assert (= ["0-1-2" "3-4-5"]
           (list-to-array
            (->> (range)
                 (take 6)
                 (partition 3)
                 (map concat-num)
                 (realize)))))

(assert (= ["0-1-2-3" "4-5"]
           (list-to-array
            (->> (range)
                 (take 6)
                 (partition 4)
                 (map concat-num)
                 (realize)))))

(assert (= ["0-1-2-3" "5"]
           (list-to-array
            (->> (range)
                 (take 6)
                 (partition #(= 4 %))
                 (map concat-num)
                 (realize)))))

(assert (= []
           (list-to-array
            (->> (range)
                 (take 6)
                 (partition 0)
                 (map concat-num)
                 (realize)))))

(assert (= ["0-1-2" "5"]
           (list-to-array
            (->> (range)
                 (take 6)
                 (partition (fn [e] (or (= e 3) (= e 4))))
                 (map concat-num)
                 (realize)))))

(assert (= ["0-1-2" "3" "4-5"]
           (list-to-array
            (->> (range)
                 (take 6)
                 (partition (fn [e] (or (= e 3) (= e 4))) true)
                 (map concat-num)
                 (realize)))))

(assert (= ["0-1-2-3" "4" "5"]
           (list-to-array
            (->> (range)
                 (take 6)
                 (partition (fn [e] (or (= e 3) (= e 4))) false)
                 (map concat-num)
                 (realize)))))

(assert (= ["0-1-2-3"  "4-5"]
           (list-to-array
            (->> (range)
                 (take 6)
                 (partition (fn [e] (or (= e 0) (= e 4))) true)
                 (map concat-num)
                 (realize)))))


;; find array index by zip
(assert (= 2 (->> (zip (range) (stream ["c" "golang" "lisp" "ruby" "python"]))
                  (filter #(= "lisp" (car (cdr %))))
                  (map #(car %))
                  (realize)
                  (car))))
(assert (= nil (->> (zip (range) (stream ["c" "golang" "lisp" "ruby" "python"]))
                  (filter #(= "c++" (car (cdr %))))
                  (map #(car %))
                  (realize)
                  (car))))

(assert (= 2 (index-of #(= % "lisp") ["c" "golang" "lisp" "ruby" "python"])))
(assert (= -1 (index-of #(= % "c++") ["c" "golang" "lisp" "ruby" "python"])))

(assert (= 2 (index-of "lisp" ["c" "golang" "lisp" "ruby" "python"])))
(assert (= -1 (index-of "c++" ["c" "golang" "lisp" "ruby" "python"])))

(assert (= "b" (nth 1 ["a" "b" "c" "d"])))
(assert (= "b" (nth 1 '("a" "b" "c" "d"))))
(assert (= 1 (nth 1 (range))))
(assert (nil? (nth -1 (range))))
(assert (nil? (nth 100 ["a" "b" "c" "d"])))


(assert (= [3 2 1] (reverse [1 2 3])))
(assert (= '(3 2 1 2) (reverse '(2 1 2 3))))

(assert (nil? (reverse nil)))
(assert (= [] (reverse [])))
(assert (= "cba" (reverse "abc")))
