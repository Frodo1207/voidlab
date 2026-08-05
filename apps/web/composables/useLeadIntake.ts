import { resolvePublicApiPath } from "../src/runtimeConfig";

type ApiEnvelope<T> = {
  code: number;
  message: string;
  data: T;
};

type LeadSubmitResponse = {
  submitted: boolean;
};

export type LeadFormInput = {
  name: string;
  contact: string;
  message: string;
};

async function requestLeadIntake<T>(path: string, payload: Record<string, unknown>) {
  const response = await fetch(resolvePublicApiPath(path), {
    method: "POST",
    headers: {
      "Content-Type": "application/json"
    },
    body: JSON.stringify(payload)
  });

  const envelope = (await response.json()) as ApiEnvelope<T>;
  if (!response.ok || envelope.code !== 0) {
    throw new Error(envelope.message || "提交失败");
  }

  return envelope.data;
}

export async function submitContactLead(input: LeadFormInput) {
  return requestLeadIntake<LeadSubmitResponse>("/api/v1/contact/submit", input);
}

export async function submitEventRsvp(eventId: number, input: LeadFormInput) {
  return requestLeadIntake<LeadSubmitResponse>(`/api/v1/events/${eventId}/rsvp`, input);
}

export async function submitBuilderInquiry(
  builderSlug: string,
  builderName: string,
  input: LeadFormInput
) {
  return requestLeadIntake<LeadSubmitResponse>("/api/v1/builders/inquiry", {
    builder_slug: builderSlug,
    builder_name: builderName,
    ...input
  });
}
