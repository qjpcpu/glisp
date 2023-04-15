(assert (= "hello world" (cdr (os/exec "echo -n hello world"))))

;; file
(def dir "./not-exist-dir/")
(def file (concat dir "not-exist-file.dat"))
(assert (not (os/file-exist? file)))
(os/write-file file "hello")
(assert (= "hello" (string (os/read-file file))))
(os/write-file file (bytes "hello"))
(assert (= "hello" (string (os/read-file file))))
(assert (os/file-exist? file))
(os/remove-file dir)
(assert (not (os/file-exist? file)))

(assert (empty? (os/env "AAAABBBBCCCC")))
(os/setenv "AAAABBBBCCCC" "111")
(assert (= "111" (os/env "AAAABBBBCCCC")))
(os/setenv "AAAABBBBCCCC" "")

(assert (not (empty? (os/read-dir "."))))
(assert (not (empty? (os/read-dir "~"))))
(assert (empty? (os/read-dir "~xxyyzz")))

(assert (= "100" (os/exec! {"cmd" "echo $XY" "env" ["XY=100"] "cwd" "/"})))
(assert (= "/" (os/exec! {"cmd" "pwd" "env" ["XY=100"] "cwd" "/"})))
