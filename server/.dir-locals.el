((go-mode . ((eval . (let ((project-root (file-name-directory
                                          (locate-dominating-file
                                           default-directory "go.mod"))))
                       (add-to-list 'dape-configs
                                    `(greener-debug
                                      modes (go-mode go-ts-mode)
                                      ensure dape-ensure-command
                                      command "dlv"
                                      command-args ("dap" "--listen" "127.0.0.1:55878")
                                      command-cwd dape-cwd-fn
                                      port 55878
                                      :type "debug"
                                      :request "launch"
                                      :mode "debug"
                                      :program ,project-root
                                      :args ["--db-url" "sqlite:///greener.db"
                                             "--auth-secret" "your-secret-key"]
                                      :cwd ,project-root)))))))
