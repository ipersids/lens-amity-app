import { PhotoIcon } from "@heroicons/react/24/outline";
import { type ChangeEventHandler, type DragEventHandler, useEffect, useMemo, useState } from "react";

const DATE_ERROR = "Date must be within the last 7 days.";

const toDateInputValue = (date: Date) => {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");

  return `${year}-${month}-${day}`;
};

const getDateRange = () => {
  const today = new Date();
  const weekAgo = new Date(today);
  weekAgo.setDate(today.getDate() - 7);

  return {
    max: toDateInputValue(today),
    min: toDateInputValue(weekAgo),
  };
};

const AddPhotoPage = () => {
  const { min: minDate, max: maxDate } = useMemo(getDateRange, []);
  const [fileName, setFileName] = useState<string>("");
  const [previewUrl, setPreviewUrl] = useState<string>("");
  const [isDragging, setIsDragging] = useState<boolean>(false);
  const [photoDate, setPhotoDate] = useState<string>(maxDate);
  const [dateError, setDateError] = useState<string>("");

  useEffect(() => {
    return () => {
      if (previewUrl) {
        URL.revokeObjectURL(previewUrl);
      }
    };
  }, [previewUrl]);

  const selectFile = (file?: File) => {
    if (file?.type.startsWith("image/")) {
      setFileName(file.name);
      setPreviewUrl(URL.createObjectURL(file));
    }
  };

  const handleFileChange: ChangeEventHandler<HTMLInputElement> = (event) => {
    selectFile(event.target.files?.[0]);
  };

  const handleDragOver: DragEventHandler<HTMLLabelElement> = (event) => {
    event.preventDefault();
    setIsDragging(true);
  };

  const handleDrop: DragEventHandler<HTMLLabelElement> = (event) => {
    event.preventDefault();
    setIsDragging(false);
    selectFile(event.dataTransfer.files[0]);
  };

  const handleDateChange: ChangeEventHandler<HTMLInputElement> = (event) => {
    const value = event.target.value;

    setPhotoDate(value);
    setDateError(value < minDate || value > maxDate ? DATE_ERROR : "");
  };

  return (
    <section className="add-photo-page">
      <form className="add-photo-form">
        <h1>Add photo</h1>

        <label
          className={`photo-upload-area${isDragging ? " dragging" : ""}`}
          onDragOver={handleDragOver}
          onDragLeave={() => setIsDragging(false)}
          onDrop={handleDrop}
        >
          {previewUrl ? (
            <img src={previewUrl} alt="" />
          ) : (
            <span className="photo-upload-icon">
              <PhotoIcon aria-hidden="true" />
            </span>
          )}

          <strong>{fileName || "Drag and drop or click to choose photo"}</strong>
          <small>{fileName ? "Click or drop another image to replace it" : "PNG or JPG, one image only"}</small>
          <input type="file" accept="image/*" onChange={handleFileChange} />
        </label>

        <label>
          Title
          <input type="text" name="title" placeholder="Morning light in Helsinki" />
        </label>

        <label>
          Short description
          <textarea name="description" placeholder="A few words about the moment" rows={4} />
        </label>

        <label>
          Date
          <input
            type="date"
            name="date"
            min={minDate}
            max={maxDate}
            value={photoDate}
            onChange={handleDateChange}
            aria-invalid={!!dateError}
            aria-describedby={dateError ? "photo-date-error" : undefined}
          />
          {dateError && (
            <span id="photo-date-error" className="auth-error">
              {dateError}
            </span>
          )}
        </label>

        <button type="button" disabled={!!dateError}>
          Add photo
        </button>
      </form>
    </section>
  );
};

export default AddPhotoPage;
