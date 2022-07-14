(defn return_deep_closure [record]
  (fn []
      (begin
        (fn [] (+  record 1))
      )
  )
)
(assert (= 1025 (((return_deep_closure 1024)))))
