(assert (exist? ["a" "b"] "b"))
(assert (not (exist? [1 2 ] 3)))
(assert (not (exist? [] 3)))

(assert (exist? '("a" "b") "b"))
(assert (not (exist? '(1 2) 3)))
(assert (not (exist? '() 3)))

(defn not-exist? [arr elem]
  (not (exist? arr elem)))

(assert (not-exist? [1 2 ] 3))
