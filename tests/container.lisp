(assert (= ["a"] (list/complement ["a" "b" "c"] ["b" "c"])))
(assert (= ["a" "c"] (list/complement ["a" "b" "c"] ["b"])))
(assert (= ["a" "b" "c"] (list/complement ["a" "b" "c"] [])))
(assert (= ["a" "b" "c"] (list/complement ["a" "b" "c"] ["d"])))
(assert (= ["c"] (list/complement ["c"] ["b"])))
(assert (= [] (list/complement [] ["b"])))

(assert (= '("a") (list/complement '("a" "b" "c") '("b" "c"))))
(assert (= '("a" "c") (list/complement '("a" "b" "c") '("b"))))
(assert (= '("a" "b" "c") (list/complement '("a" "b" "c") '())))
(assert (= '("a" "b" "c") (list/complement '("a" "b" "c") '("d"))))
(assert (= '("c") (list/complement '("c") '("b"))))
(assert (= '() (list/complement '() '("b"))))

(assert (= ["b" "c"] (list/intersect ["a" "b" "c"] ["b" "c"])))
(assert (= ["b"] (list/intersect ["a" "b" "c"] ["b"])))
(assert (= [] (list/intersect ["a" "b" "c"] [])))
(assert (= [] (list/intersect ["a" "b" "c"] ["d"])))
(assert (= [] (list/intersect ["c"] ["b"])))
(assert (= [] (list/intersect [] ["b"])))

(assert (= '("b" "c") (list/intersect '("a" "b" "c") '("b" "c"))))
(assert (= '("b") (list/intersect '("a" "b" "c") '("b"))))
(assert (= '() (list/intersect '("a" "b" "c") '())))
(assert (= '() (list/intersect '("a" "b" "c") '("d"))))
(assert (= '() (list/intersect '("c") '("b"))))
(assert (= '() (list/intersect '() '("b"))))


(assert (= ["a" "b" "c"] (uniq ["a" "a" "b" "a" "b" "c" "c" "b"])))
(assert (= '("a" "b" "c") (uniq '("a" "a" "b" "a" "b" "c" "c" "b"))))

(assert (= ["a" "b" "c"] (union ["a"] ["b" "c"])))
(assert (= '("a" "b" "c") (union '("a") '("b" "c"))))

(assert (nil? (car '())))

(assert (= '(1 2) (- '(1 2 3 4) '(3) '(4))))
(assert (= [1 2] (- [1 2 3 4] [3] [4])))

(assert (= '(1 2) (realize (- (stream '(1 2 3 4)) (stream '(3)) (stream '(4))))))
(assert (= '(1 2 3 6 7 99) (->> (- (range 1 100) (range 4 6) (range 8 99)) (realize))))

(assert (= '(1 2 3) (union '(1) '(2 3))))
(assert (= [1 2 3] (union [1] [2 3])))
(assert (= '(1 2 10 11 12) (realize (union (range 1 3) (range 10 13)))))
