import { EyeIcon, EyeSlashIcon } from "@heroicons/react/24/outline";
import { type ChangeEventHandler, useState } from "react";

type PasswordFieldProps = {
  name: string;
  autoComplete: "current-password" | "new-password";
  onChange: ChangeEventHandler<HTMLInputElement>;
  value: string;
  error?: string;
  disabled?: boolean;
  label?: string;
};

const PasswordField = ({
  name,
  autoComplete,
  onChange,
  value,
  error,
  disabled,
  label,
}: PasswordFieldProps) => {
  const [isShown, setIsShown] = useState<boolean>(false);
  const Icon = isShown ? EyeSlashIcon : EyeIcon;

  return (
    <section>
      <section className="password-field">
        <label htmlFor={autoComplete}>{label}</label>
        <input
          id={autoComplete}
          name={name}
          autoComplete={autoComplete}
          disabled={disabled}
          onChange={onChange}
          required
          type={isShown ? "text" : "password"}
          value={value}
        />
        <button
          type="button"
          onClick={() => setIsShown((current) => !current)}
          disabled={disabled}
          aria-label={isShown ? "Hide password" : "Show password"}
        >
          <Icon aria-hidden="true" />
        </button>
      </section>
      {error && <span>{error}</span>}
    </section>
  );
};

export default PasswordField;
