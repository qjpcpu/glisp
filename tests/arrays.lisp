(def testarr [1 2 3 4 5 6])

(assert (= 3 (aget testarr 2)))
(assert (= 1 (car testarr)))
(assert (= [2 3 4 5 6] (cdr testarr)))
(aset! testarr 1 0)
(assert (= [1 0 3 4 5 6] testarr))
(assert (= [3 4] (slice testarr 2 4)))
(assert (= [1 2 3] (append [1 2] 3)))
(assert (= [0 1 2 3] (concat [0 1] [2 3])))
(assert (= [0 1 2 3 4] (concat [0 1] [2 3] [4])))
(assert (= [0 1] (concat [] [] [0 1])))
(assert (= 6 (len testarr)))
(assert (= ['() '() '()] (make-array 3)))
(assert (= [0 0 0] (make-array 3 0)))

(let [a 0]
  (assert (= [a 3] (array 0 3))))

(assert (array? [1 2 3]))
(assert (empty? []))
(assert (not (empty? [1])))

(assert (= [2 3] (slice [1 2 3 4] 1 3)))
(assert (= ["b" "c"] (slice ["a" "b" "c"] 1 3)))

(assert (= 0B676c6973 (slice (str2bytes "hello glisp") 6 10)))

(assert (< [1] [1 2]))
(assert (< [1] [2]))
(assert (empty? (cdr [])))
(assert (= "a" (slice "a" 0 100)))
