(defn os/exec! [& args]
  #`Usage: (os/exec! & args)

Execute os command. Returns command output string.

examples:
(os/exec! "ls -l")
(os/exec! "ps -ef | grep glisp")
`
  (let [result (apply os/exec args)]
       (assert (= 0 (car result)) (cdr result))
       (cdr result)))
