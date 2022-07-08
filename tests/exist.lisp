(assert (exist? ["a" "b"] "b"))
(assert (not (exist? [1 2 ] 3)))
(assert (not (exist? [] 3)))

(assert (exist? '("a" "b") "b"))
(assert (not (exist? '(1 2) 3)))
(assert (not (exist? '() 3)))
