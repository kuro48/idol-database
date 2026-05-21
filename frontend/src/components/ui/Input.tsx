import { forwardRef } from 'react'
import styles from './input.module.css'

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label?: string
  error?: string
  helper?: string
}

const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ label, error, helper, id, className, ...props }, ref) => {
    const inputId = id ?? label?.toLowerCase().replace(/\s+/g, '-')

    return (
      <div className={styles.wrapper}>
        {label && (
          <label htmlFor={inputId} className={styles.label}>
            {label}
          </label>
        )}
        <input
          ref={ref}
          id={inputId}
          className={`${styles.input} ${error ? styles.inputError : ''} ${className ?? ''}`}
          aria-describedby={
            error
              ? `${inputId}-error`
              : helper
                ? `${inputId}-helper`
                : undefined
          }
          aria-invalid={error ? true : undefined}
          {...props}
        />
        {error && (
          <p id={`${inputId}-error`} className={styles.error} role="alert">
            {error}
          </p>
        )}
        {!error && helper && (
          <p id={`${inputId}-helper`} className={styles.helper}>
            {helper}
          </p>
        )}
      </div>
    )
  },
)

Input.displayName = 'Input'

export { Input }
