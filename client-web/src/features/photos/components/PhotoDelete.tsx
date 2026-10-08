import { TrashIcon } from "@heroicons/react/24/outline";

const PhotoDelete = ({ onDelete }: { onDelete: () => void }) => {
  const handleClick = (e: React.MouseEvent<HTMLButtonElement, MouseEvent>) => {
    e.stopPropagation();
    e.preventDefault();
    onDelete();
  };
  return (
    <button
      aria-label="Delete photo"
      className="photo-info-delete"
      type="button"
      onClick={(e) => handleClick(e)}
    >
      <TrashIcon aria-hidden="true" />
    </button>
  );
};

export default PhotoDelete;
