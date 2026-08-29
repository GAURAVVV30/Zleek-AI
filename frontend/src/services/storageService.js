import { apiClient } from './apiClient';
import { ENDPOINTS } from '../utils/endpoints';

/**
 * Handles Presigned URL S3 Direct Upload Flow
 * 1. Requests presigned PUT URL from backend
 * 2. Uploads binary directly to S3
 * 3. Submits reference key to backend
 */
export async function uploadProjectFile(file, conceptId, notes = '') {
  // Step 1: Request presigned URL
  const uploadConfig = await apiClient.post(ENDPOINTS.STORAGE.UPLOAD_URL, {
    filename: file.name,
    fileType: file.type || 'application/octet-stream',
    fileSize: file.size,
  });

  const { uploadUrl, fileKey } = uploadConfig.data;

  // Step 2: Direct PUT to S3 (or mock endpoint)
  if (uploadUrl) {
    await fetch(uploadUrl, {
      method: 'PUT',
      headers: {
        'Content-Type': file.type || 'application/octet-stream',
      },
      body: file,
    });
  }

  // Step 3: Submit reference
  const submissionResult = await apiClient.post(ENDPOINTS.CONCEPTS.SUBMIT_PROJECT(conceptId), {
    fileKey: fileKey || `uploads/${file.name}`,
    filename: file.name,
    fileSize: file.size,
    notes,
  });

  return submissionResult;
}
