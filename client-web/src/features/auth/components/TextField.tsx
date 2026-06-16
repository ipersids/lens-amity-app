import type { ComponentPropsWithoutRef } from "react";

type TextFieldProps = ComponentPropsWithoutRef<"input"> & {
  id: string;
  label: string;
  error?: string;
};

const TextField = ({ className, error, id, label, ...props }: TextFieldProps) => {
  const validationId = `${id}-validation`;

  return (
    <section className="auth-form-field">
      <label htmlFor={id}>{label}</label>

      <input {...props} id={id} aria-describedby={validationId} />

      {error && (
        <span id={validationId} aria-live="assertive" className="auth-error">
          {error}
        </span>
      )}
    </section>
  );
};

export default TextField;
