import { useMutation, useQueryClient } from "@tanstack/react-query";
import { ProfileQueryKey } from "../constants";
import { normalizeText } from "../features/auth/validation";
import type { UpdateProfileProps, UserProfileResponse } from "../services/users";
import usersService from "../services/users";
import { useUpdateDisplayNameInStore, useUser } from "../stores/auth";

const useUpdateProfile = () => {
  const currentUser = useUser();
  const queryClient = useQueryClient();
  const queryKey = ProfileQueryKey(currentUser?.username ?? "");
  const updateDisplayNameInStore = useUpdateDisplayNameInStore();

  return useMutation({
    mutationFn: usersService.updateUserProfile,
    onMutate: async (changes: UpdateProfileProps) => {
      await queryClient.cancelQueries({ queryKey, exact: true });

      const previousProfile = queryClient.getQueryData<UserProfileResponse>(queryKey);

      queryClient.setQueryData<UserProfileResponse>(queryKey, (profile) =>
        profile
          ? {
              ...profile,
              ...changes,
              displayName: normalizeText(changes.displayName),
              about: normalizeText(changes.about ?? ""),
            }
          : profile,
      );

      return { previousProfile };
    },

    onError: (_error, _changes, onMutateResult) => {
      if (onMutateResult?.previousProfile) {
        queryClient.setQueryData(queryKey, onMutateResult.previousProfile);
      }
    },

    onSuccess: (_data, changes) => {
      updateDisplayNameInStore(normalizeText(changes.displayName));
    },
  });
};

export default useUpdateProfile;
