import { EyeIcon, EyeSlashIcon } from "@heroicons/react/24/outline";
import { type ComponentPropsWithoutRef, useState } from "react";

type PasswordFieldProps = Omit<ComponentPropsWithoutRef<"input">, "type"> & {
  label: string;
  error?: string;
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
  ...props
}: PasswordFieldProps) => {
  const [isShown, setIsShown] = useState<boolean>(false);
  const Icon = isShown ? EyeSlashIcon : EyeIcon;
  const fieldId = id ?? name;
  const validationId = `${fieldId}-validation`;

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
          aria-describedby={validationId}
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

      {error && (
        <span id={validationId} aria-live="assertive" className="auth-error">
          {error}
        </span>
      )}
    </section>
  );
};

export default PasswordField;
