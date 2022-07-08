(defn return_deep_closure [record]
  (fn []
      (begin
        (fn [] (assert (=  record 1024)))
      )
  )
)
(((return_deep_closure 1024)))
