(assert (= 2 (len (bytes "hi"))))
(assert (bytes? (bytes "hi")))
(assert (= "hi world" (concat (bytes "hi") (bytes " ") (bytes "world"))))
