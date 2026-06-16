import { EyeIcon, EyeSlashIcon } from "@heroicons/react/24/outline";
import { type ComponentPropsWithoutRef, useState } from "react";

type PasswordFieldProps = Omit<ComponentPropsWithoutRef<"input">, "type"> & {
  label: string;
  error?: string;
  strength?: string;
};

const PasswordField = ({
  autoComplete,
  className,
  error,
  id,
  label,
  name,
  minLength = 15,
  required = true,
  strength,
  ...props
}: PasswordFieldProps) => {
  const [isShown, setIsShown] = useState<boolean>(false);
  const Icon = isShown ? EyeSlashIcon : EyeIcon;
  const fieldId = id ?? name;
  const errorId = `${fieldId}-error`;

  return (
    <section className="auth-form-field">
      <label htmlFor={fieldId}>{label}</label>

      <div className="auth-input-action">
        <input
          {...props}
          id={fieldId}
          name={name}
          type={isShown ? "text" : "password"}
          autoComplete={autoComplete}
          minLength={minLength}
          required={required}
          aria-invalid={!!error}
          aria-describedby={error ? errorId : undefined}
        />

        <button
          type="button"
          onClick={() => setIsShown((current) => !current)}
          disabled={props.disabled}
          aria-label={isShown ? "Hide password" : "Show password"}
        >
          <Icon aria-hidden="true" />
        </button>
      </div>

      {strength && <span>Strength: {strength}</span>}
      {error && (
        <span id={errorId} aria-live="assertive" className="auth-error">
          {error}
        </span>
      )}
    </section>
  );
};

export default PasswordField;
