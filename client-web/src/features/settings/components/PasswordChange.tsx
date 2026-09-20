import { XMarkIcon } from "@heroicons/react/24/outline";
import { Dialog } from "radix-ui";
import { useState } from "react";
import { useUser } from "../../../stores/auth";
import PasswordField from "../../auth/components/PasswordField";
import { validatePassword } from "../../auth/validation";

const PasswordChange = () => {
  const user = useUser();
  const [open, setOpen] = useState<boolean>(false);
  const [password, setPassword] = useState<string>("");
  const [newPassword, setNewPassword] = useState<string>("");

  if (!user) return null;

  const handleOpenChange = (isOpen: boolean) => {
    setOpen(isOpen);

    if (!isOpen) {
      setPassword("");
      setNewPassword("");
    }
  };

  const passwordValidation = validatePassword(newPassword, [user.username]);

  return (
    <Dialog.Root open={open} onOpenChange={handleOpenChange}>
      <div className="settings-account-action">
        <div>
          <h3>Change password</h3>
          <p>Update the password you use to sign in.</p>
        </div>
        <Dialog.Trigger asChild>
          <button className="dialog-trigger" type="button">
            Change password
          </button>
        </Dialog.Trigger>
        <Dialog.Portal>
          <Dialog.Overlay className="dialog-overlay" />
          <Dialog.Content className="dialog-content">
            <Dialog.Title className="dialog-title">Change password</Dialog.Title>
            <fieldset className="dialog-fieldset">
              <PasswordField
                id="current-password"
                autoComplete="current-password"
                name="current-password"
                onChange={(event) => setPassword(event.target.value)}
                value={password}
                // disabled={isLoading}
                label="Old password"
              />

              <PasswordField
                id="new-password"
                autoComplete="new-password"
                name="new-password"
                onChange={(event) => setNewPassword(event.target.value)}
                value={newPassword}
                // disabled={isLoading}
                label="New password"
                error={newPassword ? passwordValidation.feedback : undefined}
              />
            </fieldset>
            <div className="dialog-actions">
              <Dialog.Close asChild>
                <button className="dialog-cancel-button" type="button">
                  Cancel
                </button>
              </Dialog.Close>
              <Dialog.Close asChild>
                <button
                  className="dialog-save-button"
                  type="button"
                  // onClick={(e) => handleSaveChanges(e)}
                  // disabled={isPending}
                >
                  Save changes
                </button>
              </Dialog.Close>
            </div>
            <Dialog.Close asChild>
              <button className="dialog-close-button" type="button" aria-label="Close">
                <XMarkIcon />
              </button>
            </Dialog.Close>
          </Dialog.Content>
        </Dialog.Portal>
      </div>
    </Dialog.Root>
  );
};

export default PasswordChange;
