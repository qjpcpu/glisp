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


(assert (= ["a" "b" "c"] (list/uniq ["a" "a" "b" "a" "b" "c" "c" "b"])))
(assert (= '("a" "b" "c") (list/uniq '("a" "a" "b" "a" "b" "c" "c" "b"))))

(assert (= ["a" "b" "c"] (list/union ["a"] ["b" "c"])))
(assert (= '("a" "b" "c") (list/union '("a") '("b" "c"))))
