(assert (= 2 (len (str2bytes "hi"))))
(assert (bytes? (str2bytes "hi")))
(assert (= "hi world" (concat (str2bytes "hi") (str2bytes " ") (str2bytes "world"))))
