(assert (= "abc" (append "ab" #c)))
(assert (= "abcd" (concat "ab" "cd")))
(assert (= "bc" (slice "abcd" 1 3)))
