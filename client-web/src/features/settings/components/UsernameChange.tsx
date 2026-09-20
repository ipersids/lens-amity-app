import { XMarkIcon } from "@heroicons/react/24/outline";
import { Dialog } from "radix-ui";
import { useState } from "react";
import useUpdateUsername from "../../../hooks/useUpdateUsername";
import useUsernameAvailability from "../../../hooks/useUsernameAvailability";

const UsernameChange = ({ username }: { username: string }) => {
  const [open, setOpen] = useState<boolean>(false);
  const [newUsername, setNewUsername] = useState<string>("");
  const { status, statusMsg } = useUsernameAvailability({
    newUsername,
    currentUsername: username,
    delay: 500,
  });
  const updateUsername = useUpdateUsername();

  const handleOpenChange = (isOpen: boolean) => {
    setOpen(isOpen);

    if (!isOpen) {
      setNewUsername("");
    }
  };

  const handleSaveChanges = (e: React.MouseEvent<HTMLButtonElement, MouseEvent>) => {
    e.stopPropagation();
    e.preventDefault();

    updateUsername.mutate(newUsername, {
      onSuccess: () => handleOpenChange(false),
    });
  };

  return (
    <Dialog.Root open={open} onOpenChange={handleOpenChange}>
      <div className="settings-account-action">
        <div>
          <h3>Username</h3>
          <p>Change the username shown on your profile.</p>
        </div>
        <Dialog.Trigger asChild>
          <button className="dialog-trigger" type="button">
            Change username
          </button>
        </Dialog.Trigger>
        <Dialog.Portal>
          <Dialog.Overlay className="dialog-overlay" />
          <Dialog.Content className="dialog-content">
            <Dialog.Title className="dialog-title">Change username</Dialog.Title>
            <fieldset className="dialog-fieldset">
              <label className="dialog-label" htmlFor="new_username">
                Enter a new username
              </label>
              <input
                className="dialog-input"
                type="text"
                id="new_username"
                name="new_username"
                autoComplete="username"
                value={newUsername}
                onChange={(event) => {
                  event.preventDefault();
                  setNewUsername(event.target.value);
                }}
                aria-describedby="new_username_note"
                aria-invalid={
                  status === "invalid" || status === "unavailable" || status === "error"
                }
                required
              />
              <p
                className={`dialog-field-note dialog-field-note-${status}`}
                id="new_username_note"
                aria-live="polite"
              >
                {statusMsg}
              </p>
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
                  onClick={(e) => handleSaveChanges(e)}
                  disabled={status !== "available" || updateUsername.isPending}
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

export default UsernameChange;
