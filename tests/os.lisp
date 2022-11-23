(assert (= "hello world" (cdr (os/exec "echo" "-n" "hello world"))))
(assert (= "1 hello world" (cdr (os/exec "echo" "-n" 1 (bytes "hello world")))))
(def name "jack")
(assert (= "hello jack" (cdr (os/exec "echo -n hello" name))))

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
