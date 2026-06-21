(function () {
  const form = document.getElementById("upload-form");
  const fileInput = document.getElementById("file-input");
  const dropArea = document.getElementById("drop-area");
  const fileList = document.getElementById("file-list");
  const resultMessage = document.getElementById("result-message");
  const submitBtn = document.getElementById("submit-btn");
  const browseLink = dropArea.querySelector(".browse-link");

  browseLink.addEventListener("click", (e) => {
    e.preventDefault();
    fileInput.click();
  });

  dropArea.addEventListener("click", () => fileInput.click());
  fileInput.addEventListener("change", handleFiles);
  dropArea.addEventListener("dragover", (e) => {
    e.preventDefault();
    dropArea.classList.add("drag-over");
  });

  dropArea.addEventListener("dragleave", () => {
    dropArea.classList.remove("drag-over");
  });

  dropArea.addEventListener("drop", (e) => {
    e.preventDefault();
    dropArea.classList.remove("drag-over");

    const files = e.dataTransfer.files;
    fileInput.files = files;
    handleFiles();
  });

  function handleFiles() {
    const files = Array.from(fileInput.files);
    if (files.length === 0) {
      fileList.hidden = true;
      submitBtn.disabled = true;
      return;
    }

    fileList.innerHTML = "";
    files.forEach((file) => {
      const item = document.createElement("div");
      item.className = "file-item";
      item.textContent = file.name + " (" + formatFileSize(file.size) + ")";
      fileList.appendChild(item);
    });

    fileList.hidden = false;
    submitBtn.disabled = false;
    resultMessage.hidden = true;
  }

  function formatFileSize(bytes) {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  }

  form.addEventListener("submit", async (e) => {
    e.preventDefault();

    const files = fileInput.files;
    if (files.length === 0) return;

    submitBtn.disabled = true;
    submitBtn.textContent = "Scanning...";
    resultMessage.hidden = true;

    const formData = new FormData();
    Array.from(files).forEach((file) => {
      formData.append("files", file);
    });

    try {
      const response = await fetch("/scan", {
        method: "POST",
        body: formData,
      });

      const data = await response.json();

      resultMessage.className =
        "result-message " + (response.ok ? "success" : "error");
      resultMessage.textContent = response.ok
        ? "Scan complete. Total size: " + formatFileSize(data.size) + "."
        : "Error: " + data.error;
      resultMessage.hidden = false;

      if (response.ok) {
        fileInput.value = "";
        fileList.hidden = true;
        submitBtn.disabled = true;
      }
    } catch (error) {
      resultMessage.className = "result-message error";
      resultMessage.textContent = "Network error: " + error.message;
      resultMessage.hidden = false;
    } finally {
      submitBtn.textContent = "Scan Files";
      submitBtn.disabled = fileInput.files.length === 0;
    }
  });
})();
